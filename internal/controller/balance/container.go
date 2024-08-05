package balance

import (
	"github.com/m1khal3v/gophermart-loyalty-service/internal/manager"
)

type Container struct {
	userManager           *manager.UserManager
	withdrawalManager     *manager.WithdrawalManager
	userWithdrawalManager *manager.UserWithdrawalManager
}

func NewContainer(
	userManager *manager.UserManager,
	withdrawalManager *manager.WithdrawalManager,
	userWithdrawalManager *manager.UserWithdrawalManager,
) *Container {
	return &Container{
		userManager:           userManager,
		withdrawalManager:     withdrawalManager,
		userWithdrawalManager: userWithdrawalManager,
	}
}
