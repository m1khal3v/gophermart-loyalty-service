package processor

import (
	"context"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/responses"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/task"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/logger"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/manager"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/semaphore"
	"go.uber.org/zap"
)

type Updater struct {
	taskManager      *task.Manager
	orderManager     *manager.OrderManager
	userOrderManager *manager.UserOrderManager
}

func NewUpdater(taskManager *task.Manager, orderManager *manager.OrderManager, userOrderManager *manager.UserOrderManager) *Updater {
	return &Updater{
		taskManager:      taskManager,
		orderManager:     orderManager,
		userOrderManager: userOrderManager,
	}
}

func (processor *Updater) Process(ctx context.Context, concurrency uint64) error {
	semaphore := semaphore.New(concurrency)

	for {
		if err := semaphore.Acquire(ctx); err != nil {
			return err
		}

		go func() {
			defer semaphore.Release()
			if err := processor.processOne(ctx); err != nil {
				logger.Logger.Warn("can`t update order", zap.Error(err))
			}
		}()
	}
}

func (processor *Updater) processOne(ctx context.Context) error {
	accrual, ok := processor.taskManager.GetProcessed()
	if !ok {
		return nil
	}

	sum := float64(0)
	if accrual.Accrual != nil {
		sum = *accrual.Accrual
	}

	switch accrual.Status {
	case responses.AccrualStatusRegistered:
		processor.taskManager.RegisterUnprocessed(accrual.OrderID) // not final status
	case responses.AccrualStatusProcessing:
		if err := processor.orderManager.Update(ctx, accrual.OrderID, entity.OrderStatusProcessing, sum); err != nil {
			processor.taskManager.RegisterProcessed(accrual)
			return err
		}

		processor.taskManager.RegisterUnprocessed(accrual.OrderID) // not final status
	case responses.AccrualStatusInvalid:
		if err := processor.orderManager.Update(ctx, accrual.OrderID, entity.OrderStatusInvalid, sum); err != nil {
			processor.taskManager.RegisterProcessed(accrual)
			return err
		}
	case responses.AccrualStatusProcessed:
		if err := processor.userOrderManager.Accrue(ctx, accrual.OrderID, sum); err != nil {
			processor.taskManager.RegisterProcessed(accrual)
			return err
		}
	}

	return nil
}
