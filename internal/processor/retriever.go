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
	retryAfter    atomic.Pointer[time.Time]
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

		processor.processRetryAfterIfExists()

		go func() {
			defer semaphore.Release()
			if err := processor.processOne(ctx); err != nil {
				logger.Logger.Warn("can`t retrieve accrual", zap.Error(err))
			}
		}()
	}
}

func (processor *Retriever) processOne(ctx context.Context) error {
	orderID, ok := processor.unprocessed.Pop()
	if !ok {
		processor.setRetryAfter(time.Now().Add(NoTasksDelay))
		return nil
	}

	accrual, err := processor.accrualClient.GetAccrual(ctx, orderID)
	if err != nil {
		target := client.ErrTooManyRequests{}
		if errors.As(err, &target) {
			processor.setRetryAfter(target.RetryAfterTime)
			processor.unprocessed.Push(orderID)
		} else {
			processor.unprocessed.PushDelayed(ctx, orderID, FailedTaskDelay)
		}

		return fmt.Errorf("accrual %d: %w", orderID, err)
	}

	processor.processed.Push(accrual)

	return nil
}

// Lock-free processRetryAfterIfExists
func (processor *Retriever) processRetryAfterIfExists() {
	for {
		retryAfter := processor.retryAfter.Load()
		if retryAfter == nil {
			return
		}

		if sleepDuration := retryAfter.Sub(time.Now()); sleepDuration > 0 {
			time.Sleep(sleepDuration)
		}

		if processor.retryAfter.CompareAndSwap(retryAfter, nil) {
			return
		}
	}
}

// Lock-free setRetryAfter
func (processor *Retriever) setRetryAfter(retryAfter time.Time) {
	retryAfter = retryAfter.Round(time.Second)

	if retryAfter.Before(time.Now()) {
		return
	}

	for {
		if current := processor.retryAfter.Load(); current == nil || current.Before(retryAfter) {
			if !processor.retryAfter.CompareAndSwap(current, &retryAfter) {
				continue
			}
		}

		return
	}
}
