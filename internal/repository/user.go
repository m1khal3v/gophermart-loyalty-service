package repository

import (
	"context"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/gorm/repository"
	"gorm.io/gorm"
)

type UserRepository struct {
	*repository.Repository[entity.User]
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		Repository: repository.New[entity.User](db),
	}
}

func (repository *UserRepository) FindOneByLogin(ctx context.Context, login string) (*entity.User, error) {
	return repository.FindOneByField(ctx, "login", login)
}
