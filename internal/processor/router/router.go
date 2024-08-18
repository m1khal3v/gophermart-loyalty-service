package router

import (
	"context"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/responses"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/logger"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/queue"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/semaphore"
	"time"
)

const DefaultConcurrency = 10
const DefaultNoTasksDelay = time.Second * 5
const DefaultFailedTaskDelay = time.Second * 10
const DefaultNoChangesDelay = time.Minute

type Processor struct {
	orderQueue      *queue.Queue[uint64]
	routerQueue     *queue.Queue[*responses.Accrual]
	processingQueue *queue.Queue[*responses.Accrual]
	invalidQueue    *queue.Queue[*responses.Accrual]
	processedQueue  *queue.Queue[*responses.Accrual]
	config          *Config
}

type Config struct {
	Concurrency     uint64
	NoTasksDelay    *time.Duration
	FailedTaskDelay *time.Duration
	NoChangesDelay  *time.Duration
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
	if config.NoChangesDelay == nil || *config.NoChangesDelay < 0 {
		defaultValue := DefaultNoChangesDelay
		config.NoChangesDelay = &defaultValue
	}
}

func NewProcessor(
	orderQueue *queue.Queue[uint64],
	routerQueue *queue.Queue[*responses.Accrual],
	processingQueue *queue.Queue[*responses.Accrual],
	invalidQueue *queue.Queue[*responses.Accrual],
	processedQueue *queue.Queue[*responses.Accrual],
	config *Config,
) *Processor {
	prepareConfig(config)
	return &Processor{
		orderQueue:      orderQueue,
		routerQueue:     routerQueue,
		processingQueue: processingQueue,
		invalidQueue:    invalidQueue,
		processedQueue:  processedQueue,
		config:          config,
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
		processor.orderQueue.PushDelayed(ctx, accrual.OrderID, *processor.config.NoChangesDelay)
	case responses.AccrualStatusProcessing:
		processor.processingQueue.Push(accrual)
	case responses.AccrualStatusInvalid:
		processor.invalidQueue.Push(accrual)
	case responses.AccrualStatusProcessed:
		if accrual.Accrual == nil {
			processor.orderQueue.PushDelayed(ctx, accrual.OrderID, *processor.config.FailedTaskDelay)
		} else {
			processor.processedQueue.Push(accrual)
		}
	}
}

func (processor *Processor) waitIfNeed(ctx context.Context) error {
	for processor.routerQueue.Count() == 0 {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case <-time.After(*processor.config.NoTasksDelay):
			continue
		}
	}

	return nil
}
