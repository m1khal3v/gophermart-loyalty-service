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
