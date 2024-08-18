package repository

import (
	"context"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/gorm/types/money"
	"gorm.io/gorm"
)

type UserRepository struct {
	*Repository[entity.User]
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		Repository: New[entity.User](db),
	}
}

func (repository *UserRepository) FindOneByLogin(ctx context.Context, login string) (*entity.User, error) {
	return repository.FindOneBy(ctx, "login = ?", login)
}

func (repository *UserRepository) FindByID(ctx context.Context, id uint32) (*entity.User, error) {
	return repository.FindOneBy(ctx, "id = ?", id)
}

func (repository *UserRepository) Withdraw(ctx context.Context, id uint32, sum float64) (bool, error) {
	money := money.New(sum)
	affected, err := repository.Updates(ctx, &entity.User{}, map[string]interface{}{
		"balance":   gorm.Expr("balance - ?", money),
		"withdrawn": gorm.Expr("withdrawn + ?", money),
	}, "id = ? AND balance >= ?", id, money)

	if err != nil {
		return false, err
	}

	return affected == 1, nil
}

func (repository *UserRepository) Accrue(ctx context.Context, id uint32, sum float64) (bool, error) {
	money := money.New(sum)
	affected, err := repository.Updates(ctx, &entity.User{}, map[string]interface{}{
		"balance": gorm.Expr("balance + ?", money),
	}, "id = ?", id)

	if err != nil {
		return false, err
	}

	return affected == 1, nil
}
