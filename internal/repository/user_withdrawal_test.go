package repository

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/gorm/types/money"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand/v2"
	"testing"
)

func TestUserWithdrawalRepository_Withdraw(t *testing.T) {
	gorm, sqlMock := NewDBMock(t)
	repository := NewUserWithdrawalRepository(gorm)
	id := rand.Uint64N(1000) + 1
	userID := rand.Uint32N(1000) + 1
	sum := money.New(rand.Float64() + 100)

	sqlMock.ExpectBegin()
	sqlMock.
		ExpectExec(`INSERT INTO "withdrawals" ("order_id","user_id","sum","created_at") VALUES ($1,$2,$3,$4)`).
		WithArgs(int64(id), userID, sum, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(int64(id), 1))
	sqlMock.
		ExpectExec(`UPDATE "users" SET "balance"=balance - $1,"withdrawn"=withdrawn + $2,"updated_at"=$3 WHERE id = $4 AND balance >= $5`).
		WithArgs(uint64(sum), uint64(sum), sqlmock.AnyArg(), userID, uint64(sum)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	sqlMock.ExpectCommit()

	withdrawal, err := repository.Withdraw(context.Background(), id, userID, sum.AsFloat())
	require.NoError(t, err)
	assert.NotNil(t, withdrawal)
}

func TestUserWithdrawalRepository_WithdrawNoFunds(t *testing.T) {
	gorm, sqlMock := NewDBMock(t)
	repository := NewUserWithdrawalRepository(gorm)
	id := rand.Uint64N(1000) + 1
	userID := rand.Uint32N(1000) + 1
	sum := money.New(rand.Float64() + 100)

	sqlMock.ExpectBegin()
	sqlMock.
		ExpectExec(`INSERT INTO "withdrawals" ("order_id","user_id","sum","created_at") VALUES ($1,$2,$3,$4)`).
		WithArgs(int64(id), userID, sum, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(int64(id), 1))
	sqlMock.
		ExpectExec(`UPDATE "users" SET "balance"=balance - $1,"withdrawn"=withdrawn + $2,"updated_at"=$3 WHERE id = $4 AND balance >= $5`).
		WithArgs(uint64(sum), uint64(sum), sqlmock.AnyArg(), userID, uint64(sum)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	sqlMock.ExpectRollback()

	withdrawal, err := repository.Withdraw(context.Background(), id, userID, sum.AsFloat())
	require.NoError(t, err)
	assert.Nil(t, withdrawal)
}
