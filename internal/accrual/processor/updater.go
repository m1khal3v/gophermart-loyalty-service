package processor

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
)

type Updater struct {
	unprocessed      *queue.Queue[uint64]
	processed        *queue.Queue[*responses.Accrual]
	orderManager     *manager.OrderManager
	userOrderManager *manager.UserOrderManager
	concurrency      uint64
}

var ErrAccrualIsEmpty = errors.New("accrual is empty")

func NewUpdater(
	unprocessed *queue.Queue[uint64],
	processed *queue.Queue[*responses.Accrual],
	orderManager *manager.OrderManager,
	userOrderManager *manager.UserOrderManager,
	concurrency uint64,
) *Updater {
	return &Updater{
		unprocessed:      unprocessed,
		processed:        processed,
		orderManager:     orderManager,
		userOrderManager: userOrderManager,
		concurrency:      concurrency,
	}
}

func (processor *Updater) Process(ctx context.Context) error {
	semaphore := semaphore.New(processor.concurrency)

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
	accrual, ok := processor.processed.Pop()
	if !ok {
		return nil
	}

	switch accrual.Status {
	case responses.AccrualStatusRegistered:
		processor.unprocessed.Push(accrual.OrderID) // not final status
	case responses.AccrualStatusProcessing:
		if err := processor.orderManager.UpdateStatus(ctx, accrual.OrderID, entity.OrderStatusProcessing); err != nil {
			processor.processed.Push(accrual)
			return err
		}

		processor.unprocessed.Push(accrual.OrderID) // not final status
	case responses.AccrualStatusInvalid:
		if err := processor.orderManager.UpdateStatus(ctx, accrual.OrderID, entity.OrderStatusInvalid); err != nil {
			processor.processed.Push(accrual)
			return err
		}
	case responses.AccrualStatusProcessed:
		if accrual.Accrual == nil {
			processor.unprocessed.Push(accrual.OrderID)
			return ErrAccrualIsEmpty
		}

		if err := processor.userOrderManager.Accrue(ctx, accrual.OrderID, *accrual.Accrual); err != nil {
			processor.processed.Push(accrual)
			return err
		}
	}

	return nil
}
