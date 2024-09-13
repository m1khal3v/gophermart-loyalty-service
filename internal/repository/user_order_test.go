package repository

import (
	"context"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/gorm/types/money"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormerr "gorm.io/gorm"
)

func TestUserOrderRepository_AccrueOK(t *testing.T) {
	gorm, sqlMock := NewDBMock(t)
	repository := NewUserOrderRepository(gorm)
	id := rand.Uint64N(1000) + 1
	userID := rand.Uint32N(1000) + 1
	sum := money.New(rand.Float64() + 100)

	sqlMock.ExpectBegin()
	rows := sqlMock.
		NewRows([]string{"id", "user_id", "status", "accrual", "created_at", "updated_at"}).
		AddRow(int64(id), int32(userID), "TEST_STATUS", int64(sum), time.Now(), time.Now())
	sqlMock.
		ExpectQuery(`SELECT * FROM "orders" WHERE id = $1 LIMIT $2`).
		WithArgs(id, 1).
		WillReturnRows(rows)
	sqlMock.
		ExpectExec(`UPDATE "orders" SET "user_id"=$1,"status"=$2,"accrual"=$3,"created_at"=$4,"updated_at"=$5 WHERE "id" = $6`).
		WithArgs(userID, entity.OrderStatusProcessed, uint64(sum), sqlmock.AnyArg(), sqlmock.AnyArg(), id).
		WillReturnResult(sqlmock.NewResult(0, 1))
	sqlMock.
		ExpectExec(`UPDATE "users" SET "balance"=balance + $1,"updated_at"=$2 WHERE id = $3`).
		WithArgs(uint64(sum), sqlmock.AnyArg(), userID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	sqlMock.ExpectCommit()

	err := repository.Accrue(context.Background(), id, sum.AsFloat())
	require.NoError(t, err)
}

func TestUserOrderRepository_AccrueFailed(t *testing.T) {
	gorm, sqlMock := NewDBMock(t)
	repository := NewUserOrderRepository(gorm)
	id := rand.Uint64N(1000) + 1
	userID := rand.Uint32N(1000) + 1
	sum := money.New(rand.Float64() + 100)

	sqlMock.ExpectBegin()
	rows := sqlMock.
		NewRows([]string{"id", "user_id", "status", "accrual", "created_at", "updated_at"}).
		AddRow(int64(id), int32(userID), "TEST_STATUS", int64(sum), time.Now(), time.Now())
	sqlMock.
		ExpectQuery(`SELECT * FROM "orders" WHERE id = $1 LIMIT $2`).
		WithArgs(id, 1).
		WillReturnRows(rows)
	sqlMock.
		ExpectExec(`UPDATE "orders" SET "user_id"=$1,"status"=$2,"accrual"=$3,"created_at"=$4,"updated_at"=$5 WHERE "id" = $6`).
		WithArgs(userID, entity.OrderStatusProcessed, uint64(sum), sqlmock.AnyArg(), sqlmock.AnyArg(), id).
		WillReturnResult(sqlmock.NewResult(0, 1))
	sqlMock.
		ExpectExec(`UPDATE "users" SET "balance"=balance + $1,"updated_at"=$2 WHERE id = $3`).
		WithArgs(uint64(sum), sqlmock.AnyArg(), userID).
		WillReturnResult(sqlmock.NewResult(0, 0))
	sqlMock.ExpectRollback()

	err := repository.Accrue(context.Background(), id, sum.AsFloat())
	assert.ErrorIs(t, err, ErrAccrueFailed)
}

func TestUserOrderRepository_AccrueZero(t *testing.T) {
	gorm, sqlMock := NewDBMock(t)
	repository := NewUserOrderRepository(gorm)
	id := rand.Uint64N(1000) + 1
	userID := rand.Uint32N(1000) + 1
	sum := money.New(0.0)

	sqlMock.ExpectBegin()
	rows := sqlMock.
		NewRows([]string{"id", "user_id", "status", "accrual", "created_at", "updated_at"}).
		AddRow(int64(id), int32(userID), "TEST_STATUS", int64(sum), time.Now(), time.Now())
	sqlMock.
		ExpectQuery(`SELECT * FROM "orders" WHERE id = $1 LIMIT $2`).
		WithArgs(id, 1).
		WillReturnRows(rows)
	sqlMock.
		ExpectExec(`UPDATE "orders" SET "user_id"=$1,"status"=$2,"accrual"=$3,"created_at"=$4,"updated_at"=$5 WHERE "id" = $6`).
		WithArgs(userID, entity.OrderStatusProcessed, uint64(sum), sqlmock.AnyArg(), sqlmock.AnyArg(), id).
		WillReturnResult(sqlmock.NewResult(0, 1))
	sqlMock.ExpectCommit()

	err := repository.Accrue(context.Background(), id, sum.AsFloat())
	require.NoError(t, err)
}

func TestUserOrderRepository_AccrueOrderNotFound(t *testing.T) {
	gorm, sqlMock := NewDBMock(t)
	repository := NewUserOrderRepository(gorm)
	id := rand.Uint64N(1000) + 1
	sum := money.New(rand.Float64() + 100)

	sqlMock.ExpectBegin()
	sqlMock.
		ExpectQuery(`SELECT * FROM "orders" WHERE id = $1 LIMIT $2`).
		WithArgs(id, 1).
		WillReturnError(gormerr.ErrRecordNotFound)
	sqlMock.ExpectRollback()

	err := repository.Accrue(context.Background(), id, sum.AsFloat())
	require.ErrorIs(t, err, ErrOrderNotFound)
}
