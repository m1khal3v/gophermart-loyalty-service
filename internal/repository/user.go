package repository

import (
	"context"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/gorm/repository"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/gorm/types/money"
	"gorm.io/gorm"
)

type UserRepository struct {
	*repository.Repository[entity.User]
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		Repository: repository.New[entity.User](db),
		db:         db,
	}
}

func (repository *UserRepository) FindOneByLogin(ctx context.Context, login string) (*entity.User, error) {
	return repository.FindOneBy(ctx, "login = ?", login)
}

func (repository *UserRepository) FindOneByID(ctx context.Context, id uint32) (*entity.User, error) {
	return repository.FindOneBy(ctx, "id = ?", id)
}

func (repository *UserRepository) Withdraw(ctx context.Context, id uint32, sum float64) (bool, error) {
	uintSum := uint64(money.New(sum))
	result := repository.db.WithContext(ctx).Model(&entity.User{}).Where("id = ? AND balance >= ?", id, uintSum).Updates(map[string]interface{}{
		"balance":   gorm.Expr("balance - ?", uintSum),
		"withdrawn": gorm.Expr("withdrawn + ?", uintSum),
	})

	if result.Error != nil {
		return false, result.Error
	}

	return result.RowsAffected == 1, nil
}

func (repository *UserRepository) Accrue(ctx context.Context, id uint32, sum float64) (bool, error) {
	uintSum := uint64(money.New(sum))
	result := repository.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"balance": gorm.Expr("balance + ?", uintSum),
	})

	if result.Error != nil {
		return false, result.Error
	}

	return result.RowsAffected == 1, nil
}
