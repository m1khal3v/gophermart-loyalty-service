package router

import (
	"context"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/responses"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand/v2"
	"testing"
	"time"
)

func TestProcessor_processAccrualRegistered(t *testing.T) {
	orderID := rand.Uint64N(1000) + 100
	response := &responses.Accrual{
		OrderID: orderID,
		Status:  responses.AccrualStatusRegistered,
	}

	orderQueue := queue.New[uint64](1)
	routerQueue := queue.New[*responses.Accrual](1)
	invalidQueue := queue.New[*responses.Accrual](1)
	processingQueue := queue.New[*responses.Accrual](1)
	processedQueue := queue.New[*responses.Accrual](1)

	noDelay := time.Duration(0)
	processor := NewProcessor(orderQueue, routerQueue, processingQueue, invalidQueue, processedQueue, &Config{
		NoChangesDelay: &noDelay,
	})
	processor.processAccrual(context.Background(), response)

	assert.EqualValues(t, 1, orderQueue.Count())
	assert.EqualValues(t, 0, routerQueue.Count())
	assert.EqualValues(t, 0, processingQueue.Count())
	assert.EqualValues(t, 0, invalidQueue.Count())
	assert.EqualValues(t, 0, processedQueue.Count())

	retrieved, ok := orderQueue.Pop()
	require.True(t, ok)
	assert.Equal(t, response.OrderID, retrieved)
}

func TestProcessor_processAccrualProcessing(t *testing.T) {
	orderID := rand.Uint64N(1000) + 100
	response := &responses.Accrual{
		OrderID: orderID,
		Status:  responses.AccrualStatusProcessing,
	}

	orderQueue := queue.New[uint64](1)
	routerQueue := queue.New[*responses.Accrual](1)
	invalidQueue := queue.New[*responses.Accrual](1)
	processingQueue := queue.New[*responses.Accrual](1)
	processedQueue := queue.New[*responses.Accrual](1)

	noDelay := time.Duration(0)
	processor := NewProcessor(orderQueue, routerQueue, processingQueue, invalidQueue, processedQueue, &Config{
		NoChangesDelay: &noDelay,
	})
	processor.processAccrual(context.Background(), response)

	assert.EqualValues(t, 0, orderQueue.Count())
	assert.EqualValues(t, 0, routerQueue.Count())
	assert.EqualValues(t, 1, processingQueue.Count())
	assert.EqualValues(t, 0, invalidQueue.Count())
	assert.EqualValues(t, 0, processedQueue.Count())

	retrieved, ok := processingQueue.Pop()
	require.True(t, ok)
	assert.Equal(t, response, retrieved)
}

func TestProcessor_processAccrualInvalid(t *testing.T) {
	orderID := rand.Uint64N(1000) + 100
	response := &responses.Accrual{
		OrderID: orderID,
		Status:  responses.AccrualStatusInvalid,
	}

	orderQueue := queue.New[uint64](1)
	routerQueue := queue.New[*responses.Accrual](1)
	invalidQueue := queue.New[*responses.Accrual](1)
	processingQueue := queue.New[*responses.Accrual](1)
	processedQueue := queue.New[*responses.Accrual](1)

	noDelay := time.Duration(0)
	processor := NewProcessor(orderQueue, routerQueue, processingQueue, invalidQueue, processedQueue, &Config{
		NoChangesDelay: &noDelay,
	})
	processor.processAccrual(context.Background(), response)

	assert.EqualValues(t, 0, orderQueue.Count())
	assert.EqualValues(t, 0, routerQueue.Count())
	assert.EqualValues(t, 0, processingQueue.Count())
	assert.EqualValues(t, 1, invalidQueue.Count())
	assert.EqualValues(t, 0, processedQueue.Count())

	retrieved, ok := invalidQueue.Pop()
	require.True(t, ok)
	assert.Equal(t, response, retrieved)
}

func TestProcessor_processAccrualProcessedOK(t *testing.T) {
	orderID := rand.Uint64N(1000) + 100
	accrual := rand.Float64() + 100
	response := &responses.Accrual{
		OrderID: orderID,
		Status:  responses.AccrualStatusProcessed,
		Accrual: &accrual,
	}

	orderQueue := queue.New[uint64](1)
	routerQueue := queue.New[*responses.Accrual](1)
	invalidQueue := queue.New[*responses.Accrual](1)
	processingQueue := queue.New[*responses.Accrual](1)
	processedQueue := queue.New[*responses.Accrual](1)

	noDelay := time.Duration(0)
	processor := NewProcessor(orderQueue, routerQueue, processingQueue, invalidQueue, processedQueue, &Config{
		NoChangesDelay: &noDelay,
	})
	processor.processAccrual(context.Background(), response)

	assert.EqualValues(t, 0, orderQueue.Count())
	assert.EqualValues(t, 0, routerQueue.Count())
	assert.EqualValues(t, 0, processingQueue.Count())
	assert.EqualValues(t, 0, invalidQueue.Count())
	assert.EqualValues(t, 1, processedQueue.Count())

	retrieved, ok := processedQueue.Pop()
	require.True(t, ok)
	assert.Equal(t, response, retrieved)
}

func TestProcessor_processAccrualProcessedErr(t *testing.T) {
	orderID := rand.Uint64N(1000) + 100
	response := &responses.Accrual{
		OrderID: orderID,
		Status:  responses.AccrualStatusProcessed,
	}

	orderQueue := queue.New[uint64](1)
	routerQueue := queue.New[*responses.Accrual](1)
	invalidQueue := queue.New[*responses.Accrual](1)
	processingQueue := queue.New[*responses.Accrual](1)
	processedQueue := queue.New[*responses.Accrual](1)

	noDelay := time.Duration(0)
	processor := NewProcessor(orderQueue, routerQueue, processingQueue, invalidQueue, processedQueue, &Config{
		FailedTaskDelay: &noDelay,
	})
	processor.processAccrual(context.Background(), response)

	assert.EqualValues(t, 1, orderQueue.Count())
	assert.EqualValues(t, 0, routerQueue.Count())
	assert.EqualValues(t, 0, processingQueue.Count())
	assert.EqualValues(t, 0, invalidQueue.Count())
	assert.EqualValues(t, 0, processedQueue.Count())

	retrieved, ok := orderQueue.Pop()
	require.True(t, ok)
	assert.Equal(t, response.OrderID, retrieved)
}
