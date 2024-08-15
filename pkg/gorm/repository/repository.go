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

func (repository *Repository[T]) FindOneByPrimaryKey(ctx context.Context, entity *T) (*T, error) {
	result := repository.db.
		WithContext(ctx).
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

func (repository *Repository[T]) CreateOrFind(ctx context.Context, entity *T) (*T, bool, error) {
	if err := repository.Create(ctx, entity); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			entity, err := repository.FindOneByPrimaryKey(ctx, entity)
			if err != nil {
				return nil, false, err
			}

			return entity, false, nil
		}

		return nil, false, err
	}

	return entity, true, nil
}

func (repository *Repository[T]) FindBy(ctx context.Context, order, condition any, args ...any) (<-chan *T, error) {
	result, err := repository.findModelBy(ctx, new(T), order, condition, args...)
	if err != nil {
		return nil, err
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

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

func (repository *Repository[T]) FindIDsBy(ctx context.Context, order, condition any, args ...any) (<-chan uint64, error) {
	result, err := repository.findModelBy(ctx, new(T), order, condition, args...)
	if err != nil {
		return nil, err
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	return generator.NewFromFunctionWithContext(ctx, func() (uint64, bool) {
		if !result.Next() {
			return 0, false
		}

		entity := &ID{}
		if err := repository.db.ScanRows(result, entity); err != nil {
			return 0, false
		}

		return entity.ID, true
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

func (repository *Repository[T]) Save(ctx context.Context, entity *T) error {
	return repository.db.WithContext(ctx).Save(entity).Error
}

func (repository *Repository[T]) UpdateOmitZero(ctx context.Context, model *T, update *T, where any, args ...any) error {
	base := repository.db.WithContext(ctx).Model(model)
	if where != nil {
		base.Where(where, args)
	}

	return base.Updates(update).Error
}

func (repository *Repository[T]) Updates(ctx context.Context, model *T, updates any, where any, args ...any) (int64, error) {
	result := repository.db.WithContext(ctx).Model(model).Where(where, args).Updates(updates)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}
