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

func (manager *UserOrderManager) Transaction(ctx context.Context, fn func(ctx context.Context, manager *UserOrderManager) error) error {
	return manager.userOrderRepository.Transaction(ctx, func(ctx context.Context, repository *repository.UserOrderRepository) error {
		return fn(ctx, NewUserOrderManager(repository))
	})
}
