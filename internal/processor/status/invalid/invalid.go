package invalid

import (
	"context"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/responses"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/logger"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/queue"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/semaphore"
	"go.uber.org/zap"
	"time"
)

const DefaultConcurrency = 10
const DefaultBatchSize = 100
const DefaultNoTasksDelay = time.Second * 5
const DefaultFailedTaskDelay = time.Second * 10

type orderManager interface {
	UpdateStatus(ctx context.Context, ids []uint64, status string) error
}

type Processor struct {
	invalidQueue *queue.Queue[*responses.Accrual]
	orderManager orderManager
	config       *Config
}

type Config struct {
	Concurrency     uint64
	BatchSize       uint64
	NoTasksDelay    *time.Duration
	FailedTaskDelay *time.Duration
}

func prepareConfig(config *Config) {
	if config.Concurrency == 0 {
		config.Concurrency = DefaultConcurrency
	}
	if config.BatchSize == 0 {
		config.BatchSize = DefaultBatchSize
	}
	if config.NoTasksDelay == nil || *config.NoTasksDelay < 0 {
		defaultValue := DefaultNoTasksDelay
		config.NoTasksDelay = &defaultValue
	}
	if config.FailedTaskDelay == nil || *config.FailedTaskDelay < 0 {
		defaultValue := DefaultFailedTaskDelay
		config.FailedTaskDelay = &defaultValue
	}
}

func NewProcessor(
	invalidQueue *queue.Queue[*responses.Accrual],
	orderManager orderManager,
	config *Config,
) *Processor {
	prepareConfig(config)
	return &Processor{
		invalidQueue: invalidQueue,
		orderManager: orderManager,
		config:       config,
	}
}

func (processor *Processor) Process(ctx context.Context) error {
	semaphore := semaphore.New(processor.config.Concurrency)

	for {
		if err := semaphore.Acquire(ctx); err != nil {
			return err
		}

		if err := processor.waitIfNeed(ctx); err != nil {
			return err
		}

		accruals := processor.invalidQueue.PopBatch(processor.config.BatchSize)
		if len(accruals) == 0 {
			// this case should never happen
			logger.Logger.Error("accrual in invalid status queue is empty, but should not")
			semaphore.Release()
		} else {
			go func(accruals []*responses.Accrual) {
				defer semaphore.Release()
				if err := processor.processAccruals(ctx, accruals); err != nil {
					logger.Logger.Warn("can`t update orders", zap.Error(err))
				}
			}(accruals)
		}
	}
}

func (processor *Processor) processAccruals(ctx context.Context, accruals []*responses.Accrual) error {
	ids := make([]uint64, 0, len(accruals))
	for _, accrual := range accruals {
		ids = append(ids, accrual.OrderID)
	}

	if err := processor.orderManager.UpdateStatus(ctx, ids, entity.OrderStatusInvalid); err != nil {
		processor.invalidQueue.PushBatchDelayed(ctx, accruals, *processor.config.FailedTaskDelay)
		return err
	}

	return nil
}

func (processor *Processor) waitIfNeed(ctx context.Context) error {
	for processor.invalidQueue.Count() == 0 {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case <-time.After(*processor.config.NoTasksDelay):
			continue
		}
	}

	return nil
}
