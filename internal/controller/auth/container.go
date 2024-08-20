package auth

import (
	"context"
)

type userManager interface {
	Register(ctx context.Context, login, password string) (string, error)
	Authorize(ctx context.Context, login, password string) (string, error)
}

type Container struct {
	manager userManager
}

func NewContainer(manager userManager) *Container {
	return &Container{
		manager: manager,
	}
}
