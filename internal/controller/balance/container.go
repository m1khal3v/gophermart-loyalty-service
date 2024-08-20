package balance

import (
	"context"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
)

type userManager interface {
	FindByID(ctx context.Context, id uint32) (*entity.User, error)
}

type userWithdrawalManager interface {
	Withdraw(ctx context.Context, orderID uint64, userID uint32, sum float64) error
}

type Container struct {
	userManager           userManager
	userWithdrawalManager userWithdrawalManager
}

func NewContainer(
	userManager userManager,
	userWithdrawalManager userWithdrawalManager,
) *Container {
	return &Container{
		userManager:           userManager,
		userWithdrawalManager: userWithdrawalManager,
	}
}
