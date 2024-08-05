package order

import (
	accrualManager "github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/manager"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/manager"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/repository"
)

type Container struct {
	orderManager   *manager.OrderManager
	accrualManager *accrualManager.Manager
}

func NewContainer(repository *repository.OrderRepository) *Container {
	return &Container{
		orderManager:   manager.NewOrderManager(repository),
		accrualManager: accrualManager.New(repository),
	}
}
