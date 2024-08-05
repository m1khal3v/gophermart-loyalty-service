package processor

import (
	"context"
	"errors"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/client"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/manager"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/logger"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/semaphore"
	"go.uber.org/zap"
	"time"
)

type Retriever struct {
	client  client.Client
	manager manager.Manager
}

func NewRetriever(client client.Client, manager manager.Manager) *Retriever {
	return &Retriever{
		client:  client,
		manager: manager,
	}
}

func (processor *Retriever) Process(ctx context.Context, concurrency uint64) error {
	semaphore := semaphore.New(concurrency)
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
				if errors.As(err, &client.ErrTooManyRequests{}) {
					retryAfterChannel <- err.(client.ErrTooManyRequests)
				}
			}
		}()
	}
}

func (processor *Retriever) processOne(ctx context.Context) error {
	orderID, ok := processor.manager.GetUnprocessed()
	if !ok {
		return nil
	}

	accrual, err := processor.client.GetAccrual(ctx, orderID)
	if err != nil {
		return err
	}

	processor.manager.RegisterProcessed(accrual)

	return nil
}
