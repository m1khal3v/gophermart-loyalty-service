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
	taskManager  *task.Manager
	orderManager *manager.OrderManager
}

func NewUpdater(taskManager *task.Manager, orderManager *manager.OrderManager) *Updater {
	return &Updater{
		taskManager:  taskManager,
		orderManager: orderManager,
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

	if err := processor.orderManager.Update(ctx, accrual.OrderID, convertStatus(accrual.Status), sum); err != nil {
		processor.taskManager.RegisterProcessed(accrual)
		return err
	}

	if accrual.Status == responses.AccrualStatusRegistered || accrual.Status == responses.AccrualStatusProcessing {
		processor.taskManager.RegisterUnprocessed(accrual.OrderID)
	}

	return nil
}

func convertStatus(accrualStatus string) string {
	if accrualStatus == responses.AccrualStatusRegistered {
		return entity.OrderStatusNew
	}

	return accrualStatus
}
