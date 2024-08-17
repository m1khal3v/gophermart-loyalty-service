package repository

import (
	"context"
	"database/sql/driver"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/gorm/repositorytest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand/v2"
	"testing"
	"time"
)

func TestOrderRepository_FindOneByUserID(t *testing.T) {
	gorm, sqlMock := repositorytest.NewDBMock(t)
	repository := NewOrderRepository(gorm)
	id := rand.Uint64N(1000) + 1
	userID := rand.Uint32N(1000) + 1
	accrual := rand.Uint64N(1000) + 100
	rows := sqlMock.
		NewRows([]string{"id", "user_id", "status", "accrual", "created_at", "updated_at"}).
		AddRow(int64(id), int32(userID), "TEST_STATUS", int64(accrual), time.Now(), time.Now())
	sqlMock.
		ExpectQuery(`SELECT * FROM "orders" WHERE user_id = $1 LIMIT $2`).
		WithArgs(userID, 1).
		WillReturnRows(rows)

	order, err := repository.FindOneByUserID(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, id, order.ID)
	assert.Equal(t, userID, order.UserID)
	assert.Equal(t, "TEST_STATUS", order.Status)
	assert.Equal(t, accrual, uint64(order.Accrual))
}

func TestOrderRepository_FindByUserID(t *testing.T) {
	gorm, sqlMock := repositorytest.NewDBMock(t)
	repository := NewOrderRepository(gorm)
	id := rand.Uint64N(1000) + 1
	userID := rand.Uint32N(1000) + 1
	accrual := rand.Uint64N(1000) + 100
	rows := sqlMock.
		NewRows([]string{"id", "user_id", "status", "accrual", "created_at", "updated_at"}).
		AddRow(int64(id), int32(userID), "TEST_STATUS", int64(accrual), time.Now(), time.Now()).
		AddRow(int64(id)+1, int32(userID)+1, "TEST_STATUS_2", int64(accrual)+1, time.Now(), time.Now())
	sqlMock.
		ExpectQuery(`SELECT * FROM "orders" WHERE user_id = $1 ORDER BY created_at DESC`).
		WithArgs(userID).
		WillReturnRows(rows)

	orders, err := repository.FindByUserID(context.Background(), userID)
	require.NoError(t, err)
	ordersSlice := make([]*entity.Order, 0, 2)
	for order := range orders {
		ordersSlice = append(ordersSlice, order)
	}
	require.Len(t, ordersSlice, 2)

	order := ordersSlice[0]
	assert.Equal(t, id, order.ID)
	assert.Equal(t, userID, order.UserID)
	assert.Equal(t, "TEST_STATUS", order.Status)
	assert.Equal(t, accrual, uint64(order.Accrual))

	order = ordersSlice[1]
	assert.Equal(t, id+1, order.ID)
	assert.Equal(t, userID+1, order.UserID)
	assert.Equal(t, "TEST_STATUS_2", order.Status)
	assert.Equal(t, accrual+1, uint64(order.Accrual))
}

func TestOrderRepository_FindByID(t *testing.T) {
	gorm, sqlMock := repositorytest.NewDBMock(t)
	repository := NewOrderRepository(gorm)
	id := rand.Uint64N(1000) + 1
	userID := rand.Uint32N(1000) + 1
	accrual := rand.Uint64N(1000) + 100
	rows := sqlMock.
		NewRows([]string{"id", "user_id", "status", "accrual", "created_at", "updated_at"}).
		AddRow(int64(id), int32(userID), "TEST_STATUS", int64(accrual), time.Now(), time.Now())
	sqlMock.
		ExpectQuery(`SELECT * FROM "orders" WHERE id = $1 LIMIT $2`).
		WithArgs(id, 1).
		WillReturnRows(rows)

	order, err := repository.FindByID(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, id, order.ID)
	assert.Equal(t, userID, order.UserID)
	assert.Equal(t, "TEST_STATUS", order.Status)
	assert.Equal(t, accrual, uint64(order.Accrual))
}

func TestOrderRepository_FindUnprocessedIDs(t *testing.T) {
	gorm, sqlMock := repositorytest.NewDBMock(t)
	repository := NewOrderRepository(gorm)
	id := rand.Uint64N(1000) + 1
	rows := sqlMock.
		NewRows([]string{"id"}).
		AddRow(int64(id)).
		AddRow(int64(id) + 1)
	sqlMock.
		ExpectQuery(`SELECT * FROM "orders" WHERE status IN ($1,$2) ORDER BY created_at ASC`).
		WithArgs(entity.OrderStatusNew, entity.OrderStatusProcessing).
		WillReturnRows(rows)

	ids, err := repository.FindUnprocessedIDs(context.Background())
	require.NoError(t, err)
	idsSlice := make([]uint64, 0, 2)
	for queriedID := range ids {
		idsSlice = append(idsSlice, queriedID)
	}
	require.Len(t, idsSlice, 2)

	queriedID := idsSlice[0]
	assert.Equal(t, id, queriedID)

	queriedID = idsSlice[1]
	assert.Equal(t, id+1, queriedID)
}

func TestOrderRepository_UpdateStatus(t *testing.T) {
	gorm, sqlMock := repositorytest.NewDBMock(t)
	repository := NewOrderRepository(gorm)
	ids := []uint64{
		rand.Uint64N(1000) + 1,
		rand.Uint64N(1000) + 1,
		rand.Uint64N(1000) + 1,
	}

	sqlMock.ExpectBegin()
	sqlMock.
		ExpectExec(`UPDATE "orders" SET "status"=$1,"updated_at"=$2 WHERE id IN ($3,$4,$5)`).
		WithArgs("TEST_STATUS", sqlmock.AnyArg(), ids[0], ids[1], ids[2]).
		WillReturnResult(driver.ResultNoRows)
	sqlMock.ExpectCommit()

	err := repository.UpdateStatus(context.Background(), ids, "TEST_STATUS")
	require.NoError(t, err)
}
