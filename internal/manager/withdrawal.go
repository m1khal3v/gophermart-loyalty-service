package manager

import (
	"context"
	"errors"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/repository"
	"gorm.io/gorm"
)

type WithdrawalManager struct {
	withdrawalRepository *repository.WithdrawalRepository
}

func NewWithdrawalManager(withdrawalRepository *repository.WithdrawalRepository) *WithdrawalManager {
	return &WithdrawalManager{
		withdrawalRepository: withdrawalRepository,
	}
}

func (manager *WithdrawalManager) FindByUser(ctx context.Context, userID uint32) (<-chan *entity.Withdrawal, error) {
	return manager.withdrawalRepository.FindByUserID(ctx, userID)
}

func (manager *WithdrawalManager) HasUser(ctx context.Context, userID uint32) (bool, error) {
	if _, err := manager.withdrawalRepository.FindOneByUserID(ctx, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
