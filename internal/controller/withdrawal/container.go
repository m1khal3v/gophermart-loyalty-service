package withdrawal

import (
	"github.com/m1khal3v/gophermart-loyalty-service/internal/manager"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/repository"
)

type Container struct {
	manager *manager.WithdrawalManager
}

func NewContainer(repository *repository.WithdrawalRepository) *Container {
	return &Container{
		manager: manager.NewWithdrawalManager(repository),
	}
}
