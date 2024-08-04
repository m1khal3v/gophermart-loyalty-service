package balance

import (
	"github.com/m1khal3v/gophermart-loyalty-service/internal/jwt"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/manager"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/repository"
)

type Container struct {
	userManager           *manager.UserManager
	withdrawalManager     *manager.WithdrawalManager
	userWithdrawalManager *manager.UserWithdrawalManager
}

func NewContainer(
	userRepository *repository.UserRepository,
	jwt *jwt.Container,
	withdrawalRepository *repository.WithdrawalRepository,
	userWithdrawalRepository *repository.UserWithdrawalRepository,
) *Container {
	return &Container{
		userManager:           manager.NewUserManager(userRepository, jwt),
		withdrawalManager:     manager.NewWithdrawalManager(withdrawalRepository),
		userWithdrawalManager: manager.NewUserWithdrawalManager(userWithdrawalRepository),
	}
}
