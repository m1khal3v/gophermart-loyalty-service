package repository

import (
	"context"
	"database/sql"
	"errors"
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

func (repository *Repository[T]) FindOneBy(ctx context.Context, condition any, args ...any) (*T, error) {
	entity := new(T)

	result := repository.db.
		WithContext(ctx).
		Where(condition, args).
		Limit(1).
		Take(entity)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, result.Error
	}

	return entity, nil
}

func (repository *Repository[T]) FindBy(ctx context.Context, order, condition any, args ...any) (<-chan *T, error) {
	result, err := repository.findModelBy(ctx, new(T), order, condition, args...)
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

type ID struct {
	ID uint64
}

func (repository *Repository[T]) FindIDsBy(ctx context.Context, order, condition any, args ...any) (<-chan *ID, error) {
	result, err := repository.findModelBy(ctx, &ID{}, order, condition, args...)
	if err != nil {
		return nil, err
	}
	defer result.Close()

	return generator.NewFromFunctionWithContext(ctx, func() (*ID, bool) {
		if !result.Next() {
			return nil, false
		}

		entity := &ID{}
		if err := repository.db.ScanRows(result, entity); err != nil {
			return nil, false
		}

		return entity, true
	}), nil
}

func (repository *Repository[T]) findModelBy(ctx context.Context, model, order, condition any, args ...any) (*sql.Rows, error) {
	return repository.db.
		WithContext(ctx).
		Model(model).
		Where(condition, args).
		Order(order).
		Rows()
}

func (repository *Repository[T]) UpdateOmitZero(ctx context.Context, entity *T) error {
	return repository.db.WithContext(ctx).Model(new(T)).Updates(entity).Error
}
