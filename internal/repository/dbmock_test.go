package repository

import (
	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"testing"
)

func NewDBMock(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err.Error())
	}

	gorm, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{
		TranslateError: true,
	})

	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err.Error())
	}

	return gorm, mock
}
