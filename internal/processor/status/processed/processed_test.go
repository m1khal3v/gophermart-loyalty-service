package processed

import (
	"context"
	"errors"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/accrual/responses"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/queue"
	. "github.com/ovechkin-dm/mockio/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand/v2"
	"testing"
	"time"
)

func TestProcessor_processAccrualsOK(t *testing.T) {
	SetUp(t)

	count := rand.Uint64N(100) + 100
	processedQueue := queue.New[*responses.Accrual](count)
	accruals := make([]*responses.Accrual, 0, count)
	batch := make(map[uint64]float64, count)
	for i := 0; i < int(count); i++ {
		accrual := 1.11 * float64(i)
		accruals = append(accruals, &responses.Accrual{
			OrderID: uint64(i + 1),
			Status:  responses.AccrualStatusProcessed,
			Accrual: &accrual,
		})
		batch[uint64(i+1)] = accrual
	}

	userOrderManager := Mock[userOrderManager]()
	WhenSingle(userOrderManager.AccrueBatch(
		AnyContext(),
		Equal(batch),
	)).ThenReturn(nil)

	processor := NewProcessor(processedQueue, userOrderManager, &Config{})

	require.NoError(t, processor.processAccruals(context.Background(), accruals))
	assert.EqualValues(t, 0, processedQueue.Count())

	Verify(userOrderManager, Once()).AccrueBatch(
		AnyContext(),
		Equal(batch),
	)
}

func TestProcessor_processAccrualsErr(t *testing.T) {
	SetUp(t)

	count := rand.Uint64N(100) + 100
	processedQueue := queue.New[*responses.Accrual](count)
	accruals := make([]*responses.Accrual, 0, count)
	batch := make(map[uint64]float64, count)
	for i := 0; i < int(count); i++ {
		accrual := 1.11 * float64(i)
		accruals = append(accruals, &responses.Accrual{
			OrderID: uint64(i + 1),
			Status:  responses.AccrualStatusProcessed,
			Accrual: &accrual,
		})
		batch[uint64(i+1)] = accrual
	}

	userOrderManager := Mock[userOrderManager]()
	someErr := errors.New("some error")
	WhenSingle(userOrderManager.AccrueBatch(
		AnyContext(),
		Equal(batch),
	)).ThenReturn(someErr)

	noDelay := time.Duration(0)
	processor := NewProcessor(processedQueue, userOrderManager, &Config{
		FailedTaskDelay: &noDelay,
	})

	require.ErrorIs(t, processor.processAccruals(context.Background(), accruals), someErr)
	assert.EqualValues(t, count, processedQueue.Count())

	Verify(userOrderManager, Once()).AccrueBatch(
		AnyContext(),
		Equal(batch),
	)
}
