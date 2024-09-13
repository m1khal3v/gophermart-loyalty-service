package repository

import (
	"context"
	"errors"

	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/gorm/types/money"
	"gorm.io/gorm"
)

var ErrWithdrawFailed = errors.New("failed to withdraw")

type UserWithdrawalRepository struct {
	db *gorm.DB
}

func NewUserWithdrawalRepository(db *gorm.DB) *UserWithdrawalRepository {
	return &UserWithdrawalRepository{
		db: db,
	}
}

func (userWithdrawalRepository *UserWithdrawalRepository) Withdraw(ctx context.Context, orderID uint64, userID uint32, sum float64) (*entity.Withdrawal, error) {
	withdrawal := &entity.Withdrawal{
		OrderID: orderID,
		UserID:  userID,
		Sum:     money.New(sum),
	}

	err := userWithdrawalRepository.db.Transaction(func(transaction *gorm.DB) error {
		withdrawalRepository := NewWithdrawalRepository(transaction)
		userRepository := NewUserRepository(transaction)

		if err := withdrawalRepository.Create(ctx, withdrawal); err != nil {
			return err
		}

		ok, err := userRepository.Withdraw(ctx, userID, sum)
		if err != nil {
			return err
		}
		if !ok {
			return ErrWithdrawFailed
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, ErrWithdrawFailed) {
			return nil, nil
		}

		return nil, err
	}

	return withdrawal, nil
}
