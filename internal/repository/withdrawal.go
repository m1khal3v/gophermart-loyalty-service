package repository

import (
	"context"

	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
	"gorm.io/gorm"
)

type WithdrawalRepository struct {
	*Repository[entity.Withdrawal]
}

func NewWithdrawalRepository(db *gorm.DB) *WithdrawalRepository {
	return &WithdrawalRepository{
		Repository: New[entity.Withdrawal](db),
	}
}

func (repository *WithdrawalRepository) FindOneByUserID(ctx context.Context, userID uint32) (*entity.Withdrawal, error) {
	return repository.FindOneBy(ctx, "user_id = ?", userID)
}

func (repository *WithdrawalRepository) FindByUserID(ctx context.Context, userID uint32) (<-chan *entity.Withdrawal, error) {
	return repository.FindBy(ctx, "created_at DESC", "user_id = ?", userID)
}
