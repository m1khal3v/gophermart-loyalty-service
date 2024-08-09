package queue

import (
	"context"
	"time"
)

type Queue[T any] struct {
	items chan T
}

func New[T any](size uint64) *Queue[T] {
	return &Queue[T]{
		items: make(chan T, size),
	}
}

func (queue *Queue[T]) Push(item T) {
	queue.items <- item
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

func (queue *Queue[T]) Pop() (T, bool) {
	select {
	case item := <-queue.items:
		return item, true
	default:
		return *new(T), false
	}
}

func (queue *Queue[T]) Count() uint64 {
	return uint64(len(queue.items))
}
