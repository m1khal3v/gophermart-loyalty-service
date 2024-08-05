package withdrawal

import (
	"github.com/m1khal3v/gophermart-loyalty-service/internal/manager"
)

type Container struct {
	manager *manager.WithdrawalManager
}

func NewContainer(manager *manager.WithdrawalManager) *Container {
	return &Container{
		manager: manager,
	}
}
