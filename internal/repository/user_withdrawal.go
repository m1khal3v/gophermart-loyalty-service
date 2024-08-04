package repository

import (
	"context"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/gorm/types/money"
	"gorm.io/gorm"
)

type UserWithdrawalRepository struct {
	db *gorm.DB
}

func NewUserWithdrawalRepository(db *gorm.DB) *UserWithdrawalRepository {
	return &UserWithdrawalRepository{
		db: db,
	}
}

func (userWithdrawalRepository *UserWithdrawalRepository) Withdraw(ctx context.Context, orderID uint64, userID uint32, sum float64) (*entity.Withdrawal, error) {
	tx := userWithdrawalRepository.db.Begin()
	withdrawalRepository := NewWithdrawalRepository(tx)
	userRepository := NewUserRepository(tx)

	withdrawal := &entity.Withdrawal{
		OrderID: orderID,
		UserID:  userID,
		Sum:     money.New(sum),
	}

	if err := withdrawalRepository.Create(ctx, withdrawal); err != nil {
		tx.Rollback()
		return nil, err
	}

	if ok, err := userRepository.Withdraw(ctx, userID, sum); err != nil || !ok {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	return withdrawal, nil
}
