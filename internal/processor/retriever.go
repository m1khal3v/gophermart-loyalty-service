package processor

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

type Retriever struct {
	accrualClient *client.Client
	unprocessed   *queue.Queue[uint64]
	processed     *queue.Queue[*responses.Accrual]
	concurrency   uint64
	waitFor       atomic.Pointer[time.Time]
}

func NewRetriever(
	accrualClient *client.Client,
	unprocessed *queue.Queue[uint64],
	processed *queue.Queue[*responses.Accrual],
	concurrency uint64,
) *Retriever {
	return &Retriever{
		accrualClient: accrualClient,
		unprocessed:   unprocessed,
		processed:     processed,
		concurrency:   concurrency,
	}
}

func (processor *Retriever) Process(ctx context.Context) error {
	semaphore := semaphore.New(processor.concurrency)

	for {
		if err := semaphore.Acquire(ctx); err != nil {
			return err
		}

		if err := processor.waitIfNeed(ctx); err != nil {
			return err
		}

		orderID, ok := processor.unprocessed.Pop()
		if !ok {
			// this case should never happen
			logger.Logger.Error("unprocessed is empty, but should not")
			semaphore.Release()
		} else {
			go func() {
				defer semaphore.Release()
				if err := processor.processOrder(ctx, orderID); err != nil {
					logger.Logger.Warn("can`t retrieve accrual", zap.Error(err))
				}
			}()
		}
	}
}

func (processor *Retriever) processOrder(ctx context.Context, orderID uint64) error {
	accrual, err := processor.accrualClient.GetAccrual(ctx, orderID)
	if err != nil {
		target := client.ErrTooManyRequests{}
		if errors.As(err, &target) {
			processor.setWaitFor(target.RetryAfterTime)
			processor.unprocessed.Push(orderID)
		} else {
			processor.unprocessed.PushDelayed(ctx, orderID, FailedTaskDelay)
		}

		return fmt.Errorf("accrual %d: %w", orderID, err)
	}

	processor.processed.Push(accrual)

	return nil
}

// Lock-free waitIfNeed
func (processor *Retriever) waitIfNeed(ctx context.Context) error {
	for processor.unprocessed.Count() == 0 {
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

		if sleepDuration := waitFor.Sub(time.Now()); sleepDuration > 0 {
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
func (processor *Retriever) setWaitFor(new time.Time) {
	new = new.Round(time.Second)

	if new.Before(time.Now()) {
		return
	}

	for {
		if current := processor.waitFor.Load(); current == nil || current.Before(new) {
			if !processor.waitFor.CompareAndSwap(current, &new) {
				continue
			}
		}

		return
	}
}
