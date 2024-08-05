package manager

import (
	"context"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/responses"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/repository"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/queue"
)

type Manager struct {
	unprocessed *queue.Queue[uint64]
	processed   *queue.Queue[*responses.Accrual]
}

func New(repository *repository.OrderRepository) (*Manager, error) {
	unprocessedIDs, err := repository.FindUnprocessedIDs(context.TODO())
	if err != nil {
		return nil, err
	}

	unprocessedQueue := queue.New[uint64](10000)
	for id := range unprocessedIDs {
		unprocessedQueue.Push(id.ID)
	}

	return &Manager{
		unprocessed: unprocessedQueue,
		processed:   queue.New[*responses.Accrual](10000),
	}, nil
}

func (manager *Manager) RegisterUnprocessed(id uint64) {
	manager.unprocessed.Push(id)
}

func (manager *Manager) GetUnprocessed() (uint64, bool) {
	items := manager.unprocessed.Pop(1)
	if len(items) != 1 {
		return 0, false
	}

	return items[0], true
}

func (manager *Manager) RegisterProcessed(accrual *responses.Accrual) {
	manager.processed.Push(accrual)
}

func (manager *Manager) GetProcessed() (*responses.Accrual, bool) {
	items := manager.processed.Pop(1)
	if len(items) != 1 {
		return nil, false
	}

	return items[0], true
}
