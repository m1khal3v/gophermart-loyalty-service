package repository

import (
	"context"
	"errors"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/gorm/repository"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/gorm/types/money"
	"gorm.io/gorm"
)

type OrderRepository struct {
	*repository.Repository[entity.Order]
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{
		Repository: repository.New[entity.Order](db),
	}
}

func (repository *OrderRepository) CreateOrFind(ctx context.Context, order *entity.Order) (*entity.Order, bool, error) {
	if err := repository.Create(ctx, order); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			order, err := repository.FindOneBy(ctx, "id = ?", order.ID)
			if err != nil {
				return nil, false, err
			}

			return order, false, nil
		}

		return nil, false, err
	}

	return order, true, nil
}

func (repository *OrderRepository) FindOneByUserID(ctx context.Context, userID uint32) (*entity.Order, error) {
	return repository.FindOneBy(ctx, "user_id = ?", userID)
}

func (repository *OrderRepository) FindByUserID(ctx context.Context, userID uint32) (<-chan *entity.Order, error) {
	return repository.FindBy(ctx, "created_at DESC", "user_id = ?", userID)
}

func (repository *OrderRepository) FindUnprocessedIDs(ctx context.Context) (<-chan *repository.ID, error) {
	return repository.FindIDsBy(ctx, "created_at ASC", "status IN (?,?)", entity.OrderStatusNew, entity.OrderStatusProcessing)
}

func (repository *OrderRepository) UpdateByID(ctx context.Context, id uint64, status string, accrual float64) error {
	return repository.UpdateOmitZero(ctx, &entity.Order{
		ID:      id,
		Status:  status,
		Accrual: money.New(accrual),
	})
}
