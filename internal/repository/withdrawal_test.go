package repository

import (
	"context"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/gorm/repositorytest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand/v2"
	"testing"
	"time"
)

func TestWithdrawalRepository_FindOneByUserID(t *testing.T) {
	gorm, sqlMock := repositorytest.NewDBMock(t)
	repository := NewWithdrawalRepository(gorm)
	id := rand.Uint64N(1000) + 1
	userID := rand.Uint32N(1000) + 1
	sum := rand.Uint64N(1000) + 100
	rows := sqlMock.
		NewRows([]string{"order_id", "user_id", "sum", "created_at"}).
		AddRow(int64(id), int32(userID), int64(sum), time.Now())
	sqlMock.
		ExpectQuery(`SELECT * FROM "withdrawals" WHERE user_id = $1 LIMIT $2`).
		WithArgs(userID, 1).
		WillReturnRows(rows)

	withdrawal, err := repository.FindOneByUserID(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, id, withdrawal.OrderID)
	assert.Equal(t, userID, withdrawal.UserID)
	assert.Equal(t, sum, uint64(withdrawal.Sum))
}

func TestWithdrawalRepository_FindByUserID(t *testing.T) {
	gorm, sqlMock := repositorytest.NewDBMock(t)
	repository := NewWithdrawalRepository(gorm)
	id := rand.Uint64N(1000) + 1
	userID := rand.Uint32N(1000) + 1
	sum := rand.Uint64N(1000) + 100
	rows := sqlMock.
		NewRows([]string{"order_id", "user_id", "sum", "created_at"}).
		AddRow(int64(id), int32(userID), int64(sum), time.Now()).
		AddRow(int64(id)+1, int32(userID)+1, int64(sum)+1, time.Now())
	sqlMock.
		ExpectQuery(`SELECT * FROM "withdrawals" WHERE user_id = $1 ORDER BY created_at DESC`).
		WithArgs(userID).
		WillReturnRows(rows)

	withdrawals, err := repository.FindByUserID(context.Background(), userID)
	require.NoError(t, err)
	withdrawalsSlice := make([]*entity.Withdrawal, 0, 2)
	for withdrawal := range withdrawals {
		withdrawalsSlice = append(withdrawalsSlice, withdrawal)
	}
	require.Len(t, withdrawalsSlice, 2)

	withdrawal := withdrawalsSlice[0]
	assert.Equal(t, id, withdrawal.OrderID)
	assert.Equal(t, userID, withdrawal.UserID)
	assert.Equal(t, sum, uint64(withdrawal.Sum))

	withdrawal = withdrawalsSlice[1]
	assert.Equal(t, id+1, withdrawal.OrderID)
	assert.Equal(t, userID+1, withdrawal.UserID)
	assert.Equal(t, sum+1, uint64(withdrawal.Sum))
}
