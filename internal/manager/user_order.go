package manager

import (
	"context"

	"github.com/m1khal3v/gophermart-loyalty-service/internal/repository"
)

type UserOrderManager struct {
	userOrderRepository *repository.UserOrderRepository
}

func NewUserOrderManager(userOrderRepository *repository.UserOrderRepository) *UserOrderManager {
	return &UserOrderManager{
		userOrderRepository: userOrderRepository,
	}
}

func (manager *UserOrderManager) Accrue(ctx context.Context, orderID uint64, accrual float64) error {
	return manager.userOrderRepository.Accrue(ctx, orderID, accrual)
}

func (manager *UserOrderManager) AccrueBatch(ctx context.Context, accruals map[uint64]float64) error {
	return manager.userOrderRepository.Transaction(ctx, func(ctx context.Context, repository *repository.UserOrderRepository) error {
		for orderID, accrual := range accruals {
			if err := repository.Accrue(ctx, orderID, accrual); err != nil {
				return err
			}
		}

		return nil
	})
}
