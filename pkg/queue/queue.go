package queue

import (
	"context"
	"time"
)

type Queue[T any] struct {
	items chan T
}

type removeBatchFilter[T any] func(items []T) error

func New[T any](size uint64) *Queue[T] {
	return &Queue[T]{
		items: make(chan T, size),
	}
}

func (queue *Queue[T]) Push(item T) {
	queue.items <- item
}

func (queue *Queue[T]) PushBatch(items []T) {
	for _, item := range items {
		queue.items <- item
	}
}

func (queue *Queue[T]) PushChannel(items <-chan T) {
	for item := range items {
		queue.items <- item
	}
}

func (queue *Queue[T]) PushDelayed(ctx context.Context, item T, delay time.Duration) {
	go func() {
		select {
		case <-ctx.Done():
			// push cancelled if context cancelled
		case <-time.After(delay):
			queue.items <- item
		}
	}()
}

func (queue *Queue[T]) PushBatchDelayed(ctx context.Context, items []T, delay time.Duration) {
	for _, item := range items {
		queue.PushDelayed(ctx, item, delay)
	}
}

func (queue *Queue[T]) Pop() (T, bool) {
	select {
	case item := <-queue.items:
		return item, true
	default:
		return *new(T), false
	}
}

func (queue *Queue[T]) PopBatch(count uint64) []T {
	if count == 0 || len(queue.items) == 0 {
		return []T{}
	}

	items := make([]T, 0, count)

	for i := uint64(0); i < count; i++ {
		select {
		case item := <-queue.items:
			items = append(items, item)
		default:
			// no more items to return
			return items
		}
	}

	return items
}

func (queue *Queue[T]) RemoveBatch(count uint64, filter removeBatchFilter[T]) error {
	items := queue.PopBatch(count)
	if err := filter(items); err != nil {
		queue.PushBatch(items)

		return err
	}

	return nil
}

func (queue *Queue[T]) Count() uint64 {
	return uint64(len(queue.items))
}
