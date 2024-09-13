package manager

import (
	"context"

	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
)

type withdrawalRepository interface {
	FindOneByUserID(ctx context.Context, userID uint32) (*entity.Withdrawal, error)
	FindByUserID(ctx context.Context, userID uint32) (<-chan *entity.Withdrawal, error)
}

type WithdrawalManager struct {
	withdrawalRepository withdrawalRepository
}

func NewWithdrawalManager(withdrawalRepository withdrawalRepository) *WithdrawalManager {
	return &WithdrawalManager{
		withdrawalRepository: withdrawalRepository,
	}
}

func (manager *WithdrawalManager) FindByUser(ctx context.Context, userID uint32) (<-chan *entity.Withdrawal, error) {
	return manager.withdrawalRepository.FindByUserID(ctx, userID)
}

func (manager *WithdrawalManager) HasUser(ctx context.Context, userID uint32) (bool, error) {
	withdrawal, err := manager.withdrawalRepository.FindOneByUserID(ctx, userID)
	if err != nil {
		return false, err
	}

	return withdrawal != nil, nil
}
