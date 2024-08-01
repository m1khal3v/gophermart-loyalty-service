package user

import (
	"github.com/m1khal3v/gophermart-loyalty-service/internal/manager"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/repository"
)

type Container struct {
	manager *manager.UserManager
}

func NewContainer(repository *repository.UserRepository) *Container {
	return &Container{
		manager: manager.NewUserManager(repository),
	}
}
