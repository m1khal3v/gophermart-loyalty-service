package repository

import (
	"context"
	"errors"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/gorm/types/money"
	"gorm.io/gorm"
)

var ErrOrderNotFound = errors.New("order not found")
var ErrAccrueFailed = errors.New("failed to accrue")

type UserOrderRepository struct {
	db *gorm.DB
}

func NewUserOrderRepository(db *gorm.DB) *UserOrderRepository {
	return &UserOrderRepository{
		db: db,
	}
}

func (userOrderRepository *UserOrderRepository) Accrue(ctx context.Context, orderID uint64, accrual float64) error {
	return userOrderRepository.db.Transaction(func(transaction *gorm.DB) error {
		orderRepository := NewOrderRepository(transaction)
		userRepository := NewUserRepository(transaction)

		order, err := orderRepository.FindByID(ctx, orderID)
		if err != nil {
			return err
		}
		if order == nil {
			return ErrOrderNotFound
		}

		order.Status = entity.OrderStatusProcessed
		order.Accrual = money.New(accrual)
		if err := orderRepository.Save(ctx, order); err != nil {
			return err
		}
		if accrual == 0 {
			return nil
		}

		ok, err := userRepository.Accrue(ctx, order.UserID, accrual)
		if err != nil {
			return err
		}
		if !ok {
			return ErrAccrueFailed
		}

		return nil
	})
}

func (userOrderRepository *UserOrderRepository) Transaction(ctx context.Context, fn func(ctx context.Context, repository *UserOrderRepository) error) error {
	return userOrderRepository.db.Transaction(func(transaction *gorm.DB) error {
		return fn(ctx, NewUserOrderRepository(transaction))
	})
}
