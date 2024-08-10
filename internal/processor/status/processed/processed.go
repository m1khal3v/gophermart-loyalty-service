package processed

import (
	"context"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/responses"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/logger"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/manager"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/queue"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/semaphore"
	"go.uber.org/zap"
	"time"
)

const NoTasksDelay = time.Second * 5
const FailedTaskDelay = time.Second * 10

type Processor struct {
	processedQueue   *queue.Queue[*responses.Accrual]
	userOrderManager *manager.UserOrderManager
	concurrency      uint64
	batchSize        uint64
}

func NewProcessor(
	processedQueue *queue.Queue[*responses.Accrual],
	userOrderManager *manager.UserOrderManager,
	concurrency uint64,
	batchSize uint64,
) *Processor {
	return &Processor{
		processedQueue:   processedQueue,
		userOrderManager: userOrderManager,
		concurrency:      concurrency,
		batchSize:        batchSize,
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

		accruals := processor.processedQueue.PopBatch(processor.batchSize)
		if len(accruals) == 0 {
			// this case should never happen
			logger.Logger.Error("accrual in processed status queue is empty, but should not")
			semaphore.Release()
		} else {
			go func() {
				defer semaphore.Release()
				if err := processor.processAccruals(ctx, accruals); err != nil {
					logger.Logger.Warn("can`t update orders", zap.Error(err))
				}
			}()
		}
	}
}

func (processor *Processor) processAccruals(ctx context.Context, accruals []*responses.Accrual) error {
	err := processor.userOrderManager.Transaction(ctx, func(ctx context.Context, manager *manager.UserOrderManager) error {
		for _, accrual := range accruals {
			if err := manager.Accrue(ctx, accrual.OrderID, *accrual.Accrual); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		for _, accrual := range accruals {
			processor.processedQueue.PushDelayed(ctx, accrual, FailedTaskDelay)
		}

		return err
	}

	return nil
}

// Lock-free waitIfNeed
func (processor *Processor) waitIfNeed(ctx context.Context) error {
	for processor.processedQueue.Count() == 0 {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case <-time.After(NoTasksDelay):
			continue
		}
	}

	return nil
}
