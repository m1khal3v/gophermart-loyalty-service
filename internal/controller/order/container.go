package order

import (
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/task"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/manager"
)

type Container struct {
	orderManager *manager.OrderManager
	taskManager  *task.Manager
}

func NewContainer(orderManager *manager.OrderManager, taskManager *task.Manager) *Container {
	return &Container{
		orderManager: orderManager,
		taskManager:  taskManager,
	}
}
