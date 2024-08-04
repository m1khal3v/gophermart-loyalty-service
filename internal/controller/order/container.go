package order

import (
	"github.com/m1khal3v/gophermart-loyalty-service/internal/manager"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/repository"
)

type Container struct {
	manager *manager.OrderManager
}

func NewContainer(repository *repository.OrderRepository) *Container {
	return &Container{
		manager: manager.NewOrderManager(repository),
	}
}
