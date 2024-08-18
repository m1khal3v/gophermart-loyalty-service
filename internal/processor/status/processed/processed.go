package processed

import (
	"context"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/responses"
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

type userOrderManager interface {
	AccrueBatch(ctx context.Context, accruals map[uint64]float64) error
}

type Processor struct {
	processedQueue   *queue.Queue[*responses.Accrual]
	userOrderManager userOrderManager
	config           *Config
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
	processedQueue *queue.Queue[*responses.Accrual],
	userOrderManager userOrderManager,
	config *Config,
) *Processor {
	prepareConfig(config)
	return &Processor{
		processedQueue:   processedQueue,
		userOrderManager: userOrderManager,
		config:           config,
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

		accruals := processor.processedQueue.PopBatch(processor.config.BatchSize)
		if len(accruals) == 0 {
			// this case should never happen
			logger.Logger.Error("accrual in processed status queue is empty, but should not")
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
	batch := make(map[uint64]float64, len(accruals))
	for _, accrual := range accruals {
		batch[accrual.OrderID] = *accrual.Accrual
	}

	if err := processor.userOrderManager.AccrueBatch(ctx, batch); err != nil {
		processor.processedQueue.PushBatchDelayed(ctx, accruals, *processor.config.FailedTaskDelay)
		return err
	}

	return nil
}

func (processor *Processor) waitIfNeed(ctx context.Context) error {
	for processor.processedQueue.Count() == 0 {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case <-time.After(*processor.config.NoTasksDelay):
			continue
		}
	}

	return nil
}
