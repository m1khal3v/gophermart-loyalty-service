package retriever

import (
	"context"
	"errors"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/client"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/responses"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/queue"
	. "github.com/ovechkin-dm/mockio/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand/v2"
	"testing"
	"time"
)

func TestProcessor_processOrderOK(t *testing.T) {
	SetUp(t)

	orderID := rand.Uint64N(1000) + 100
	accrual := rand.Float64() + 100
	response := &responses.Accrual{
		OrderID: orderID,
		Status:  "TEST_STATUS",
		Accrual: &accrual,
	}

	accrualClient := Mock[accrualClient]()
	WhenDouble(accrualClient.GetAccrual(
		AnyContext(),
		Exact(orderID),
	)).ThenReturn(response, nil)
	orderQueue := queue.New[uint64](1)
	accrualQueue := queue.New[*responses.Accrual](1)

	processor := NewProcessor(accrualClient, orderQueue, accrualQueue, &Config{})
	err := processor.processOrder(context.Background(), orderID)
	Verify(accrualClient, Once()).GetAccrual(
		AnyContext(),
		Exact(orderID),
	)

	require.NoError(t, err)
	assert.EqualValues(t, 0, orderQueue.Count())
	assert.EqualValues(t, 1, accrualQueue.Count())
	retrieved, ok := accrualQueue.Pop()
	require.True(t, ok)
	assert.Equal(t, response, retrieved)
	assert.Nil(t, processor.waitFor.Load())
}

func TestProcessor_processOrderErr(t *testing.T) {
	SetUp(t)

	orderID := rand.Uint64N(1000) + 100
	someErr := errors.New("some error")

	accrualClient := Mock[accrualClient]()
	WhenDouble(accrualClient.GetAccrual(
		AnyContext(),
		Exact(orderID),
	)).ThenReturn(nil, someErr)
	orderQueue := queue.New[uint64](1)
	accrualQueue := queue.New[*responses.Accrual](1)

	noDelay := time.Duration(0)
	processor := NewProcessor(accrualClient, orderQueue, accrualQueue, &Config{
		FailedTaskDelay: &noDelay,
	})

	err := processor.processOrder(context.Background(), orderID)
	Verify(accrualClient, Once()).GetAccrual(
		AnyContext(),
		Exact(orderID),
	)

	require.ErrorIs(t, err, someErr)
	assert.EqualValues(t, 1, orderQueue.Count())
	assert.EqualValues(t, 0, accrualQueue.Count())
	retrieved, ok := orderQueue.Pop()
	require.True(t, ok)
	assert.Equal(t, orderID, retrieved)
	assert.Nil(t, processor.waitFor.Load())
}

func TestProcessor_processOrderTooManyRequests(t *testing.T) {
	SetUp(t)

	orderID := rand.Uint64N(1000) + 100
	someErr := client.ErrTooManyRequests{RetryAfterTime: time.Now().Add(time.Hour)}

	accrualClient := Mock[accrualClient]()
	WhenDouble(accrualClient.GetAccrual(
		AnyContext(),
		Exact(orderID),
	)).ThenReturn(nil, someErr)
	orderQueue := queue.New[uint64](1)
	accrualQueue := queue.New[*responses.Accrual](1)

	noDelay := time.Duration(0)
	processor := NewProcessor(accrualClient, orderQueue, accrualQueue, &Config{
		FailedTaskDelay: &noDelay,
	})

	err := processor.processOrder(context.Background(), orderID)
	Verify(accrualClient, Once()).GetAccrual(
		AnyContext(),
		Exact(orderID),
	)

	require.ErrorIs(t, err, someErr)
	assert.EqualValues(t, 1, orderQueue.Count())
	assert.EqualValues(t, 0, accrualQueue.Count())
	retrieved, ok := orderQueue.Pop()
	require.True(t, ok)
	assert.Equal(t, orderID, retrieved)
	assert.NotNil(t, processor.waitFor.Load())
}
