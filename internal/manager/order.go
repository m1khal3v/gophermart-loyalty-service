package manager

import (
	"context"
	"errors"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/repository"
	"gorm.io/gorm"
)

var ErrOrderAlreadyRegisteredByCurrentUser = errors.New("order already registered by current user")
var ErrOrderAlreadyRegisteredByAnotherUser = errors.New("order already registered by another user")

type OrderManager struct {
	orderRepository *repository.OrderRepository
}

func NewOrderManager(orderRepository *repository.OrderRepository) *OrderManager {
	return &OrderManager{
		orderRepository: orderRepository,
	}
}

func (manager *OrderManager) Register(ctx context.Context, id uint64, userID uint32) (*entity.Order, error) {
	order := &entity.Order{
		ID:     id,
		UserID: userID,
		Status: entity.OrderStatusNew,
	}

	order, created, err := manager.orderRepository.CreateOrFind(ctx, order)
	if err != nil {
		return nil, err
	}
	if created {
		return order, nil
	}

	if order.UserID == userID {
		return order, ErrOrderAlreadyRegisteredByCurrentUser
	}

	return order, ErrOrderAlreadyRegisteredByAnotherUser
}

func (manager *OrderManager) FindByUser(ctx context.Context, userID uint32) (<-chan *entity.Order, error) {
	return manager.orderRepository.FindByUserID(ctx, userID)
}

func (manager *OrderManager) HasUser(ctx context.Context, userID uint32) (bool, error) {
	if _, err := manager.orderRepository.FindOneByUserID(ctx, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (manager *OrderManager) UpdateStatus(ctx context.Context, ids []uint64, status string) error {
	return manager.orderRepository.UpdateStatus(ctx, ids, status)
}
