package queue

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
