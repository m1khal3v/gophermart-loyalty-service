package updater

import (
	"context"
	"errors"
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
const NotFinalStatusDelay = time.Minute

type Processor struct {
	orderQueue       *queue.Queue[uint64]
	accrualQueue     *queue.Queue[*responses.Accrual]
	orderManager     *manager.OrderManager
	userOrderManager *manager.UserOrderManager
	concurrency      uint64
}

var ErrAccrualIsEmpty = errors.New("accrual is empty")

func New(
	orderQueue *queue.Queue[uint64],
	accrualQueue *queue.Queue[*responses.Accrual],
	orderManager *manager.OrderManager,
	userOrderManager *manager.UserOrderManager,
	concurrency uint64,
) *Processor {
	return &Processor{
		orderQueue:       orderQueue,
		accrualQueue:     accrualQueue,
		orderManager:     orderManager,
		userOrderManager: userOrderManager,
		concurrency:      concurrency,
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

		accrual, ok := processor.accrualQueue.Pop()
		if !ok {
			// this case should never happen
			logger.Logger.Error("accrual queue is empty, but should not")
			semaphore.Release()
		} else {
			go func() {
				defer semaphore.Release()
				if err := processor.processAccrual(ctx, accrual); err != nil {
					logger.Logger.Warn("can`t update order", zap.Error(err))
				}
			}()
		}
	}
}

func (processor *Processor) processAccrual(ctx context.Context, accrual *responses.Accrual) error {
	switch accrual.Status {
	case responses.AccrualStatusRegistered:
		processor.orderQueue.PushDelayed(ctx, accrual.OrderID, NotFinalStatusDelay)
	case responses.AccrualStatusProcessing:
		if err := processor.orderManager.UpdateStatus(ctx, accrual.OrderID, entity.OrderStatusProcessing); err != nil {
			processor.accrualQueue.PushDelayed(ctx, accrual, FailedTaskDelay)
			return err
		}

		processor.orderQueue.PushDelayed(ctx, accrual.OrderID, NotFinalStatusDelay)
	case responses.AccrualStatusInvalid:
		if err := processor.orderManager.UpdateStatus(ctx, accrual.OrderID, entity.OrderStatusInvalid); err != nil {
			processor.accrualQueue.PushDelayed(ctx, accrual, FailedTaskDelay)
			return err
		}
	case responses.AccrualStatusProcessed:
		if accrual.Accrual == nil {
			processor.orderQueue.PushDelayed(ctx, accrual.OrderID, FailedTaskDelay)
			return ErrAccrualIsEmpty
		}

		if err := processor.userOrderManager.Accrue(ctx, accrual.OrderID, *accrual.Accrual); err != nil {
			processor.accrualQueue.PushDelayed(ctx, accrual, FailedTaskDelay)
			return err
		}
	}

	return nil
}

// Lock-free waitIfNeed
func (processor *Processor) waitIfNeed(ctx context.Context) error {
	for processor.accrualQueue.Count() == 0 {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case <-time.After(NoTasksDelay):
			continue
		}
	}

	return nil
}
