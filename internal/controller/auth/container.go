package auth

import (
	"context"
)

type UserManager interface {
	Register(ctx context.Context, login, password string) (string, error)
	Authorize(ctx context.Context, login, password string) (string, error)
}

type Container struct {
	manager UserManager
}

func NewContainer(manager UserManager) *Container {
	return &Container{
		manager: manager,
	}
}
