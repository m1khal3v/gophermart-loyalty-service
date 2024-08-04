package repository

import (
	"context"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/gorm/repository"
	"gorm.io/gorm"
)

type WithdrawalRepository struct {
	*repository.Repository[entity.Withdrawal]
}

func NewWithdrawalRepository(db *gorm.DB) *WithdrawalRepository {
	return &WithdrawalRepository{
		Repository: repository.New[entity.Withdrawal](db),
	}
}

func (repository *WithdrawalRepository) FindOneByUserID(ctx context.Context, userID uint32) (*entity.Withdrawal, error) {
	return repository.FindOneByField(ctx, "user_id", userID)
}

func (repository *WithdrawalRepository) FindByUserID(ctx context.Context, userID uint32) (<-chan *entity.Withdrawal, error) {
	return repository.FindByField(ctx, "user_id", userID, "created_at DESC")
}
