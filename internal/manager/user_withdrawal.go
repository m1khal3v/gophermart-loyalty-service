package manager

import (
	"context"
	"errors"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/repository"
)

var ErrInsufficientFunds = errors.New("insufficient funds")

type UserWithdrawalManager struct {
	userWithdrawalRepository *repository.UserWithdrawalRepository
}

func NewUserWithdrawalManager(userWithdrawalRepository *repository.UserWithdrawalRepository) *UserWithdrawalManager {
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
