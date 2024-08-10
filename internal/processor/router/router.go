package router

import (
	"context"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/responses"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/logger"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/queue"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/semaphore"
	"time"
)

const NoTasksDelay = time.Second * 5
const FailedTaskDelay = time.Second * 10
const NoChangesDelay = time.Minute

type Processor struct {
	orderQueue      *queue.Queue[uint64]
	routerQueue     *queue.Queue[*responses.Accrual]
	processingQueue *queue.Queue[*responses.Accrual]
	invalidQueue    *queue.Queue[*responses.Accrual]
	processedQueue  *queue.Queue[*responses.Accrual]
	concurrency     uint64
}

func NewProcessor(
	orderQueue *queue.Queue[uint64],
	routerQueue *queue.Queue[*responses.Accrual],
	processingQueue *queue.Queue[*responses.Accrual],
	invalidQueue *queue.Queue[*responses.Accrual],
	processedQueue *queue.Queue[*responses.Accrual],
	concurrency uint64,
) *Processor {
	return &Processor{
		orderQueue:      orderQueue,
		routerQueue:     routerQueue,
		processingQueue: processingQueue,
		invalidQueue:    invalidQueue,
		processedQueue:  processedQueue,
		concurrency:     concurrency,
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

		accrual, ok := processor.routerQueue.Pop()
		if !ok {
			// this case should never happen
			logger.Logger.Error("router queue is empty, but should not")
			semaphore.Release()
		} else {
			go func(accrual *responses.Accrual) {
				defer semaphore.Release()
				processor.processAccrual(ctx, accrual)
			}(accrual)
		}
	}
}

func (processor *Processor) processAccrual(ctx context.Context, accrual *responses.Accrual) {
	switch accrual.Status {
	case responses.AccrualStatusRegistered:
		processor.orderQueue.PushDelayed(ctx, accrual.OrderID, NoChangesDelay)
	case responses.AccrualStatusProcessing:
		processor.processingQueue.Push(accrual)
	case responses.AccrualStatusInvalid:
		processor.invalidQueue.Push(accrual)
	case responses.AccrualStatusProcessed:
		if accrual.Accrual == nil {
			processor.orderQueue.PushDelayed(ctx, accrual.OrderID, FailedTaskDelay)
		} else {
			processor.processedQueue.Push(accrual)
		}
	}
}

// Lock-free waitIfNeed
func (processor *Processor) waitIfNeed(ctx context.Context) error {
	for processor.routerQueue.Count() == 0 {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case <-time.After(NoTasksDelay):
			continue
		}
	}

	return nil
}
