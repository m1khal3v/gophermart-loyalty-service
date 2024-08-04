package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/generator"
	"gorm.io/gorm"
)

type Repository[T any] struct {
	db *gorm.DB
}

func New[T any](db *gorm.DB) *Repository[T] {
	return &Repository[T]{db: db}
}

func (repository *Repository[T]) Create(ctx context.Context, entity *T) error {
	return repository.db.WithContext(ctx).Create(entity).Error
}

func (repository *Repository[T]) FindOneByField(ctx context.Context, field string, value any) (*T, error) {
	entity := new(T)

	result := repository.db.WithContext(ctx).Where(fmt.Sprintf("%s = ?", field), value).Limit(1).Take(entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, result.Error
	}

	return entity, nil
}

func (repository *Repository[T]) FindByField(ctx context.Context, field string, value, order any) (<-chan *T, error) {
	entity := new(T)

	result, err := repository.db.WithContext(ctx).Model(entity).Where(fmt.Sprintf("%s = ?", field), value).Order(order).Rows()
	if err != nil {
		return nil, err
	}
	defer result.Close()

	return generator.NewFromFunctionWithContext(ctx, func() (*T, bool) {
		if !result.Next() {
			return nil, false
		}

		entity := new(T)
		if err := repository.db.ScanRows(result, entity); err != nil {
			return nil, false
		}

		return entity, true
	}), nil
}
