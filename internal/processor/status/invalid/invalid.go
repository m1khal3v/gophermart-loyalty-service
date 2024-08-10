package invalid

import (
	"context"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/responses"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
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
	invalidQueue *queue.Queue[*responses.Accrual]
	orderManager *manager.OrderManager
	concurrency  uint64
	batchSize    uint64
}

func NewProcessor(
	invalidQueue *queue.Queue[*responses.Accrual],
	orderManager *manager.OrderManager,
	concurrency uint64,
	batchSize uint64,
) *Processor {
	return &Processor{
		invalidQueue: invalidQueue,
		orderManager: orderManager,
		concurrency:  concurrency,
		batchSize:    batchSize,
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

		accruals := processor.invalidQueue.PopBatch(processor.batchSize)
		if len(accruals) == 0 {
			// this case should never happen
			logger.Logger.Error("accrual in invalid status queue is empty, but should not")
			semaphore.Release()
		} else {
			go func(accruals []*responses.Accrual) {
				defer semaphore.Release()
				if err := processor.processAccruals(ctx, accruals); err != nil {
					logger.Logger.Warn("can`t update orders", zap.Error(err))
				}
			}(accruals)
		}
	}
}

func (processor *Processor) processAccruals(ctx context.Context, accruals []*responses.Accrual) error {
	ids := make([]uint64, len(accruals))
	for _, accrual := range accruals {
		ids = append(ids, accrual.OrderID)
	}

	if err := processor.orderManager.UpdateStatus(ctx, ids, entity.OrderStatusInvalid); err != nil {
		for _, accrual := range accruals {
			processor.invalidQueue.PushDelayed(ctx, accrual, FailedTaskDelay)
		}

		return err
	}

	return nil
}

// Lock-free waitIfNeed
func (processor *Processor) waitIfNeed(ctx context.Context) error {
	for processor.invalidQueue.Count() == 0 {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case <-time.After(NoTasksDelay):
			continue
		}
	}

	return nil
}
