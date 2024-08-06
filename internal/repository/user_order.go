package repository

import (
	"context"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/gorm/types/money"
	"gorm.io/gorm"
)

type UserOrderRepository struct {
	db *gorm.DB
}

func NewUserOrderRepository(db *gorm.DB) *UserOrderRepository {
	return &UserOrderRepository{
		db: db,
	}
}

func (userOrderRepository *UserOrderRepository) Accrue(ctx context.Context, orderID uint64, accrual float64) error {
	tx := userOrderRepository.db.Begin()
	orderRepository := NewOrderRepository(tx)
	userRepository := NewUserRepository(tx)

	order, err := orderRepository.FindByID(ctx, orderID)
	if err != nil {
		tx.Rollback()
		return err
	}

	order.Status = entity.OrderStatusProcessed
	order.Accrual = money.New(accrual)
	if err := orderRepository.Save(ctx, order); err != nil {
		tx.Rollback()
		return err
	}

	if ok, err := userRepository.Accrue(ctx, order.UserID, accrual); err != nil || !ok {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}
