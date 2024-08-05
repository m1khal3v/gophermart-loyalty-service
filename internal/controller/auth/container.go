package auth

import (
	"github.com/m1khal3v/gophermart-loyalty-service/internal/manager"
)

type Container struct {
	manager *manager.UserManager
}

func NewContainer(manager *manager.UserManager) *Container {
	return &Container{
		manager: manager,
	}
}
