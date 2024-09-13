package manager

import (
	"context"
	"errors"

	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
)

var ErrInsufficientFunds = errors.New("insufficient funds")

type userWithdrawalRepository interface {
	Withdraw(ctx context.Context, orderID uint64, userID uint32, sum float64) (*entity.Withdrawal, error)
}

type UserWithdrawalManager struct {
	userWithdrawalRepository userWithdrawalRepository
}

func NewUserWithdrawalManager(userWithdrawalRepository userWithdrawalRepository) *UserWithdrawalManager {
	return &UserWithdrawalManager{
		userWithdrawalRepository: userWithdrawalRepository,
	}
}

func (manager *UserWithdrawalManager) Withdraw(ctx context.Context, orderID uint64, userID uint32, sum float64) error {
	withdrawal, err := manager.userWithdrawalRepository.Withdraw(ctx, orderID, userID, sum)
	if err != nil {
		return err
	}
	if withdrawal == nil {
		return ErrInsufficientFunds
	}

	return nil
}
