package order

import (
	"github.com/m1khal3v/gophermart-loyalty-service/internal/manager"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/queue"
)

type Container struct {
	orderManager     *manager.OrderManager
	unprocessedQueue *queue.Queue[uint64]
}

func NewContainer(orderManager *manager.OrderManager, unprocessedQueue *queue.Queue[uint64]) *Container {
	return &Container{
		orderManager:     orderManager,
		unprocessedQueue: unprocessedQueue,
	}
}
