package generator

import (
	"context"
	"golang.org/x/exp/maps"
	"sync"
)

type keyValueModifier[K comparable, T any] func(key K, value T) (K, T)
type valueModifier[T any] func(value T) T

func NewFromFunction[T any](generate func() (T, bool)) <-chan T {
	return NewFromFunctionWithContext[T](context.Background(), generate)
}

func NewFromFunctionWithContext[T any](ctx context.Context, generate func() (T, bool)) <-chan T {
	if generate == nil {
		panic("generate function cannot be nil")
	}

	channel := make(chan T, 1)

	go func() {
		defer close(channel)

		for {
			value, ok := generate()
			if !ok {
				return
			}

			select {
			case <-ctx.Done():
				return
			default:
				channel <- value
			}
		}
	}()

	return channel
}

type mapItem[K comparable, T any] struct {
	Key   K
	Value T
}

func NewFromMap[K comparable, T any](source map[K]T, modify keyValueModifier[K, T]) <-chan mapItem[K, T] {
	return NewFromMapWithContext[K, T](context.Background(), source, modify)
}

func NewFromMapWithContext[K comparable, T any](
	ctx context.Context,
	source map[K]T,
	modify keyValueModifier[K, T],
) <-chan mapItem[K, T] {
	channel := make(chan mapItem[K, T], 1)

	go func() {
		defer close(channel)

		for key, value := range maps.Clone(source) {
			if modify != nil {
				key, value = modify(key, value)
			}

			select {
			case <-ctx.Done():
				return
			default:
				channel <- mapItem[K, T]{key, value}
			}
		}
	}()

	return channel
}

func NewFromSyncMap[K comparable, T any](source *sync.Map, modify keyValueModifier[K, T]) <-chan mapItem[K, T] {
	return NewFromSyncMapWithContext[K, T](context.Background(), source, modify)
}

func NewFromSyncMapWithContext[K comparable, T any](
	ctx context.Context,
	source *sync.Map,
	modify keyValueModifier[K, T],
) <-chan mapItem[K, T] {
	channel := make(chan mapItem[K, T], 1)

	go func() {
		defer close(channel)

		source.Range(func(key, value any) bool {
			keyK, valueT := key.(K), value.(T)
			if modify != nil {
				keyK, valueT = modify(keyK, valueT)
			}

			select {
			case <-ctx.Done():
				return false
			default:
				channel <- mapItem[K, T]{keyK, valueT}
				return true
			}
		})
	}()

	return channel
}

func NewFromMapOnlyValue[K comparable, T any](source map[K]T, modify valueModifier[T]) <-chan T {
	return NewFromMapOnlyValueWithContext[K, T](context.Background(), source, modify)
}

func NewFromMapOnlyValueWithContext[K comparable, T any](
	ctx context.Context,
	source map[K]T,
	modify valueModifier[T],
) <-chan T {
	channel := make(chan T, 1)

	go func() {
		defer close(channel)

		for _, value := range maps.Clone(source) {
			if modify != nil {
				value = modify(value)
			}

			select {
			case <-ctx.Done():
				return
			default:
				channel <- value
			}
		}
	}()

	return channel
}

func NewFromSyncMapOnlyValue[T any](source *sync.Map, modify valueModifier[T]) <-chan T {
	return NewFromSyncMapOnlyValueWithContext[T](context.Background(), source, modify)
}

func NewFromSyncMapOnlyValueWithContext[T any](
	ctx context.Context,
	source *sync.Map,
	modify valueModifier[T],
) <-chan T {
	channel := make(chan T, 1)

	go func() {
		defer close(channel)

		source.Range(func(_, value any) bool {
			valueT := value.(T)
			if modify != nil {
				valueT = modify(valueT)
			}

			select {
			case <-ctx.Done():
				return false
			default:
				channel <- valueT
				return true
			}
		})
	}()

	return channel
}
