package processor

import (
	"context"
	"errors"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/client"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/responses"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/logger"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/queue"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/semaphore"
	"go.uber.org/zap"
	"time"
)

type Retriever struct {
	accrualClient *client.Client
	unprocessed   *queue.Queue[uint64]
	processed     *queue.Queue[*responses.Accrual]
	concurrency   uint64
}

func NewRetriever(
	accrualClient *client.Client,
	unprocessed *queue.Queue[uint64],
	processed *queue.Queue[*responses.Accrual],
	concurrency uint64,
) *Retriever {
	return &Retriever{
		accrualClient: accrualClient,
		unprocessed:   unprocessed,
		processed:     processed,
		concurrency:   concurrency,
	}
}

func (processor *Retriever) Process(ctx context.Context) error {
	semaphore := semaphore.New(processor.concurrency)
	retryAfterChannel := make(chan client.ErrTooManyRequests, 1)
	defer close(retryAfterChannel)

	for {
		select {
		case err := <-retryAfterChannel:
			now := time.Now()
			if now.Before(err.RetryAfterTime) {
				time.Sleep(err.RetryAfterTime.Sub(now))
			}
			continue
		default:
		}

		if err := semaphore.Acquire(ctx); err != nil {
			return err
		}

		go func() {
			defer semaphore.Release()
			if err := processor.processOne(ctx); err != nil {
				logger.Logger.Warn("can`t retrieve accrual", zap.Error(err))
				target := client.ErrTooManyRequests{}
				if errors.As(err, &target) {
					retryAfterChannel <- target
				}
			}
		}()
	}
}

func (processor *Retriever) processOne(ctx context.Context) error {
	orderID, ok := processor.unprocessed.Pop()
	if !ok {
		return nil
	}

	accrual, err := processor.accrualClient.GetAccrual(ctx, orderID)
	if err != nil {
		return err
	}

	processor.processed.Push(accrual)

	return nil
}
