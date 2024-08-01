package repository

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
)

type Repository[T any] struct {
	db *gorm.DB
}

func New[T any](db *gorm.DB) *Repository[T] {
	return &Repository[T]{db: db}
}

func (repository *Repository[T]) Create(entity *T) error {
	return repository.db.Create(entity).Error
}

func (repository *Repository[T]) FindOneByField(field string, value any) (*T, error) {
	entity := new(T)

	result := repository.db.Where(fmt.Sprintf("%s = ?", field), value).Limit(1).Take(entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, result.Error
	}

	return entity, nil
}
