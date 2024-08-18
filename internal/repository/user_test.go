package repository

import (
	"context"
	"database/sql/driver"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/gorm/types/money"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand/v2"
	"testing"
	"time"
)

func TestUserRepository_FindOneByLogin(t *testing.T) {
	gorm, sqlMock := NewDBMock(t)
	repository := NewUserRepository(gorm)
	id := rand.Uint32N(1000) + 1
	balance := rand.Uint64N(1000) + 100
	withdrawn := balance - 1
	rows := sqlMock.
		NewRows([]string{"id", "login", "password", "balance", "withdrawn", "created_at", "updated_at"}).
		AddRow(int32(id), "test_login", []byte{}, balance, withdrawn, time.Now(), time.Now())
	sqlMock.
		ExpectQuery(`SELECT * FROM "users" WHERE login = $1 LIMIT $2`).
		WithArgs("test_login", 1).
		WillReturnRows(rows)

	user, err := repository.FindOneByLogin(context.Background(), "test_login")
	require.NoError(t, err)
	assert.Equal(t, id, user.ID)
	assert.Equal(t, "test_login", user.Login)
	assert.Equal(t, balance, uint64(user.Balance))
	assert.Equal(t, withdrawn, uint64(user.Withdrawn))
}

func TestUserRepository_FindByID(t *testing.T) {
	gorm, sqlMock := NewDBMock(t)
	repository := NewUserRepository(gorm)
	id := rand.Uint32N(1000) + 1
	balance := rand.Uint64N(1000) + 100
	withdrawn := balance - 1
	rows := sqlMock.
		NewRows([]string{"id", "login", "password", "balance", "withdrawn", "created_at", "updated_at"}).
		AddRow(int32(id), "test_login", []byte{}, balance, withdrawn, time.Now(), time.Now())
	sqlMock.
		ExpectQuery(`SELECT * FROM "users" WHERE id = $1 LIMIT $2`).
		WithArgs(id, 1).
		WillReturnRows(rows)

	user, err := repository.FindByID(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, id, user.ID)
	assert.Equal(t, "test_login", user.Login)
	assert.Equal(t, balance, uint64(user.Balance))
	assert.Equal(t, withdrawn, uint64(user.Withdrawn))
}

func TestUserRepository_Withdraw(t *testing.T) {
	tests := []struct {
		name   string
		result driver.Result
		ok     bool
	}{
		{
			name:   "success",
			result: sqlmock.NewResult(0, 1),
			ok:     true,
		},
		{
			name:   "failure",
			result: driver.ResultNoRows,
			ok:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gorm, sqlMock := NewDBMock(t)
			repository := NewUserRepository(gorm)
			id := rand.Uint32N(1000) + 1
			sum := money.New(rand.Float64() + 100)

			sqlMock.ExpectBegin()
			sqlMock.
				ExpectExec(`UPDATE "users" SET "balance"=balance - $1,"withdrawn"=withdrawn + $2,"updated_at"=$3 WHERE id = $4 AND balance >= $5`).
				WithArgs(uint64(sum), uint64(sum), sqlmock.AnyArg(), id, uint64(sum)).
				WillReturnResult(tt.result)
			sqlMock.ExpectCommit()

			ok, err := repository.Withdraw(context.Background(), id, sum.AsFloat())
			require.NoError(t, err)
			assert.Equal(t, tt.ok, ok)
		})
	}
}

func TestUserRepository_Accrue(t *testing.T) {
	tests := []struct {
		name   string
		result driver.Result
		ok     bool
	}{
		{
			name:   "success",
			result: sqlmock.NewResult(0, 1),
			ok:     true,
		},
		{
			name:   "failure",
			result: driver.ResultNoRows,
			ok:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gorm, sqlMock := NewDBMock(t)
			repository := NewUserRepository(gorm)
			id := rand.Uint32N(1000) + 1
			sum := money.New(rand.Float64() + 100)

			sqlMock.ExpectBegin()
			sqlMock.
				ExpectExec(`UPDATE "users" SET "balance"=balance + $1,"updated_at"=$2 WHERE id = $3`).
				WithArgs(uint64(sum), sqlmock.AnyArg(), id).
				WillReturnResult(tt.result)
			sqlMock.ExpectCommit()

			ok, err := repository.Accrue(context.Background(), id, sum.AsFloat())
			require.NoError(t, err)
			assert.Equal(t, tt.ok, ok)
		})
	}
}
