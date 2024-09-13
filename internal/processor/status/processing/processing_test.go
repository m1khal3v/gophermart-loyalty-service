package processing

import (
	"context"
	"errors"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/responses"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/queue"
	. "github.com/ovechkin-dm/mockio/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessor_processAccrualsOK(t *testing.T) {
	SetUp(t)

	count := rand.Uint64N(100) + 100
	orderQueue := queue.New[uint64](count)
	processingQueue := queue.New[*responses.Accrual](count)
	accruals := make([]*responses.Accrual, 0, count)
	ids := make([]uint64, 0, count)
	for i := 0; i < int(count); i++ {
		accruals = append(accruals, &responses.Accrual{
			OrderID: uint64(i + 1),
			Status:  responses.AccrualStatusProcessing,
		})
		ids = append(ids, uint64(i+1))
	}
	orderManager := Mock[orderManager]()
	WhenSingle(orderManager.UpdateStatus(
		AnyContext(),
		Equal(ids),
		Exact(responses.AccrualStatusProcessing),
	)).ThenReturn(nil)

	noDelay := time.Duration(0)
	processor := NewProcessor(orderQueue, processingQueue, orderManager, &Config{
		NotFinalStatusDelay: &noDelay,
	})

	require.NoError(t, processor.processAccruals(context.Background(), accruals))
	assert.EqualValues(t, 0, processingQueue.Count())
	assert.EqualValues(t, count, orderQueue.Count())

	Verify(orderManager, Once()).UpdateStatus(
		AnyContext(),
		Equal(ids),
		Exact(responses.AccrualStatusProcessing),
	)
}

func TestProcessor_processAccrualsErr(t *testing.T) {
	SetUp(t)

	count := rand.Uint64N(100) + 100
	orderQueue := queue.New[uint64](count)
	processingQueue := queue.New[*responses.Accrual](count)
	accruals := make([]*responses.Accrual, 0, count)
	ids := make([]uint64, 0, count)
	for i := 0; i < int(count); i++ {
		accruals = append(accruals, &responses.Accrual{
			OrderID: uint64(i + 1),
			Status:  responses.AccrualStatusProcessing,
		})
		ids = append(ids, uint64(i+1))
	}
	someErr := errors.New("some error")
	orderManager := Mock[orderManager]()
	WhenSingle(orderManager.UpdateStatus(
		AnyContext(),
		Equal(ids),
		Exact(responses.AccrualStatusProcessing),
	)).ThenReturn(someErr)

	noDelay := time.Duration(0)
	processor := NewProcessor(orderQueue, processingQueue, orderManager, &Config{
		FailedTaskDelay: &noDelay,
	})

	require.ErrorIs(t, processor.processAccruals(context.Background(), accruals), someErr)
	assert.EqualValues(t, count, processingQueue.Count())
	assert.EqualValues(t, 0, orderQueue.Count())

	Verify(orderManager, Once()).UpdateStatus(
		AnyContext(),
		Equal(ids),
		Exact(responses.AccrualStatusProcessing),
	)
}
