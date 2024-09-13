package retriever

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/client"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/responses"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/logger"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/queue"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/semaphore"
	"go.uber.org/zap"
)

const DefaultConcurrency = 10
const DefaultNoTasksDelay = time.Second * 5
const DefaultFailedTaskDelay = time.Second * 10

type accrualClient interface {
	GetAccrual(ctx context.Context, orderID uint64) (*responses.Accrual, error)
}

type Processor struct {
	accrualClient accrualClient
	orderQueue    *queue.Queue[uint64]
	accrualQueue  *queue.Queue[*responses.Accrual]
	waitFor       atomic.Pointer[time.Time]
	config        *Config
}

type Config struct {
	Concurrency     uint64
	NoTasksDelay    *time.Duration
	FailedTaskDelay *time.Duration
}

func prepareConfig(config *Config) {
	if config.Concurrency == 0 {
		config.Concurrency = DefaultConcurrency
	}
	if config.NoTasksDelay == nil || *config.NoTasksDelay < 0 {
		defaultValue := DefaultNoTasksDelay
		config.NoTasksDelay = &defaultValue
	}
	if config.FailedTaskDelay == nil || *config.FailedTaskDelay < 0 {
		defaultValue := DefaultFailedTaskDelay
		config.FailedTaskDelay = &defaultValue
	}
}

func NewProcessor(
	accrualClient accrualClient,
	orderQueue *queue.Queue[uint64],
	accrualQueue *queue.Queue[*responses.Accrual],
	config *Config,
) *Processor {
	prepareConfig(config)
	return &Processor{
		accrualClient: accrualClient,
		orderQueue:    orderQueue,
		accrualQueue:  accrualQueue,
		config:        config,
	}
}

func (processor *Processor) Process(ctx context.Context) error {
	semaphore := semaphore.New(processor.config.Concurrency)

	for {
		if err := semaphore.Acquire(ctx); err != nil {
			return err
		}

		if err := processor.waitIfNeed(ctx); err != nil {
			return err
		}

		orderID, ok := processor.orderQueue.Pop()
		if !ok {
			// this case should never happen
			logger.Logger.Error("order queue is empty, but should not")
			semaphore.Release()
		} else {
			go func(orderID uint64) {
				defer semaphore.Release()
				if err := processor.processOrder(ctx, orderID); err != nil {
					logger.Logger.Warn("can`t retrieve accrual", zap.Error(err))
				}
			}(orderID)
		}
	}
}

func (processor *Processor) processOrder(ctx context.Context, orderID uint64) error {
	accrual, err := processor.accrualClient.GetAccrual(ctx, orderID)
	if err != nil {
		target := client.ErrTooManyRequests{}
		if errors.As(err, &target) {
			processor.setWaitFor(target.RetryAfterTime)
			processor.orderQueue.Push(orderID)
		} else {
			processor.orderQueue.PushDelayed(ctx, orderID, *processor.config.FailedTaskDelay)
		}

		return fmt.Errorf("accrual %d: %w", orderID, err)
	}

	processor.accrualQueue.Push(accrual)

	return nil
}

// Lock-free waitIfNeed
func (processor *Processor) waitIfNeed(ctx context.Context) error {
	for processor.orderQueue.Count() == 0 {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case <-time.After(*processor.config.NoTasksDelay):
			continue
		}
	}

	for {
		waitFor := processor.waitFor.Load()
		if waitFor == nil {
			return nil
		}

		if sleepDuration := time.Until(*waitFor); sleepDuration > 0 {
			select {
			case <-ctx.Done():
				return context.Cause(ctx)
			case <-time.After(sleepDuration):
				// sleep done
			}
		}

		if processor.waitFor.CompareAndSwap(waitFor, nil) {
			return nil
		}
	}
}

// Lock-free setWaitFor
func (processor *Processor) setWaitFor(new time.Time) {
	new = new.Round(time.Second)

	for {
		if new.Before(time.Now()) {
			return
		}

		if current := processor.waitFor.Load(); current == nil || current.Before(new) {
			if !processor.waitFor.CompareAndSwap(current, &new) {
				continue
			}
		}

		return
	}
}
