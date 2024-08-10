package retriever

import (
	"context"
	"errors"
	"fmt"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/client"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/responses"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/logger"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/queue"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/semaphore"
	"go.uber.org/zap"
	"sync/atomic"
	"time"
)

const NoTasksDelay = time.Second * 5
const FailedTaskDelay = time.Second * 10

type Processor struct {
	accrualClient *client.Client
	orderQueue    *queue.Queue[uint64]
	accrualQueue  *queue.Queue[*responses.Accrual]
	concurrency   uint64
	waitFor       atomic.Pointer[time.Time]
}

func NewProcessor(
	accrualClient *client.Client,
	orderQueue *queue.Queue[uint64],
	accrualQueue *queue.Queue[*responses.Accrual],
	concurrency uint64,
) *Processor {
	return &Processor{
		accrualClient: accrualClient,
		orderQueue:    orderQueue,
		accrualQueue:  accrualQueue,
		concurrency:   concurrency,
	}
}

func (processor *Processor) Process(ctx context.Context) error {
	semaphore := semaphore.New(processor.concurrency)

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
			processor.orderQueue.PushDelayed(ctx, orderID, FailedTaskDelay)
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
		case <-time.After(NoTasksDelay):
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
