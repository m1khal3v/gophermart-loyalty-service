package withdrawal

import (
	"context"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
)

type withdrawalManager interface {
	FindByUser(ctx context.Context, userID uint32) (<-chan *entity.Withdrawal, error)
	HasUser(ctx context.Context, userID uint32) (bool, error)
}

type Container struct {
	manager withdrawalManager
}

func NewContainer(manager withdrawalManager) *Container {
	return &Container{
		manager: manager,
	}
}
