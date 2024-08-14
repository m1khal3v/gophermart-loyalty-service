package order

import (
	"context"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/queue"
)

type orderManager interface {
	Register(ctx context.Context, id uint64, userID uint32) (*entity.Order, error)
	FindByUser(ctx context.Context, userID uint32) (<-chan *entity.Order, error)
	HasUser(ctx context.Context, userID uint32) (bool, error)
}

type Container struct {
	orderManager orderManager
	orderQueue   *queue.Queue[uint64]
}

func NewContainer(orderManager orderManager, orderQueue *queue.Queue[uint64]) *Container {
	return &Container{
		orderManager: orderManager,
		orderQueue:   orderQueue,
	}
}
