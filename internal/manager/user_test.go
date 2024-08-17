package manager

import (
	"context"
	"errors"
	"fmt"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/jwt"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/gorm/types/bcrypt"
	. "github.com/ovechkin-dm/mockio/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"testing"
)

func TestUserManager_Register(t *testing.T) {
	someErr := errors.New("some error")
	tests := []struct {
		name       string
		repository func() userRepository
		verify     func(repository userRepository)
		wantToken  bool
		wantErr    error
	}{
		{
			name: "ok",
			repository: func() userRepository {
				repository := Mock[userRepository]()
				WhenSingle(repository.Create(
					AnyContext(),
					Match(CreateMatcher("provided user", func(allArgs []any, actual *entity.User) bool {
						return actual.Login == "login_1" &&
							actual.Password.CompareWithPassword("password_1") == nil
					})),
				)).ThenReturn(nil)

				return repository
			},
			verify: func(repository userRepository) {
				Verify(repository, Once()).Create(
					AnyContext(),
					Match(CreateMatcher("provided user", func(allArgs []any, actual *entity.User) bool {
						return actual.Login == "login_1" &&
							actual.Password.CompareWithPassword("password_1") == nil
					})),
				)
			},
			wantToken: true,
		},
		{
			name: "login already exists",
			repository: func() userRepository {
				repository := Mock[userRepository]()
				WhenSingle(repository.Create(
					AnyContext(),
					Match(CreateMatcher("provided user", func(allArgs []any, actual *entity.User) bool {
						return actual.Login == "login_2" &&
							actual.Password.CompareWithPassword("password_2") == nil
					})),
				)).ThenReturn(gorm.ErrDuplicatedKey)

				return repository
			},
			verify: func(repository userRepository) {
				Verify(repository, Once()).Create(
					AnyContext(),
					Match(CreateMatcher("provided user", func(allArgs []any, actual *entity.User) bool {
						return actual.Login == "login_2" &&
							actual.Password.CompareWithPassword("password_2") == nil
					})),
				)
			},
			wantToken: false,
			wantErr:   ErrLoginAlreadyExists,
		},
		{
			name: "db error",
			repository: func() userRepository {
				repository := Mock[userRepository]()
				WhenSingle(repository.Create(
					AnyContext(),
					Match(CreateMatcher("provided user", func(allArgs []any, actual *entity.User) bool {
						return actual.Login == "login_3" &&
							actual.Password.CompareWithPassword("password_3") == nil
					})),
				)).ThenReturn(someErr)

				return repository
			},
			verify: func(repository userRepository) {
				Verify(repository, Once()).Create(
					AnyContext(),
					Match(CreateMatcher("provided user", func(allArgs []any, actual *entity.User) bool {
						return actual.Login == "login_3" &&
							actual.Password.CompareWithPassword("password_3") == nil
					})),
				)
			},
			wantToken: false,
			wantErr:   someErr,
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetUp(t)
			i++
			repository := tt.repository()
			jwt := jwt.New(fmt.Sprintf("secret_%d", i))
			manager := NewUserManager(repository, jwt)

			token, err := manager.Register(context.Background(), fmt.Sprintf("login_%d", i), fmt.Sprintf("password_%d", i))
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			if tt.wantToken {
				assert.NotEmpty(t, token)
				claims, err := jwt.Decode(token)
				require.NoError(t, err)
				assert.Equal(t, fmt.Sprintf("login_%d", i), claims.Subject)
			} else {
				assert.Empty(t, token)
			}

			tt.verify(repository)
		})
	}
}

func TestUserManager_Authorize(t *testing.T) {
	someErr := errors.New("some error")
	tests := []struct {
		name       string
		repository func() userRepository
		verify     func(repository userRepository)
		wantToken  bool
		wantErr    error
	}{
		{
			name: "ok",
			repository: func() userRepository {
				repository := Mock[userRepository]()
				password, _ := bcrypt.NewHash("password_1", 1)
				WhenDouble(repository.FindOneByLogin(
					AnyContext(),
					Exact("login_1"),
				)).ThenReturn(&entity.User{
					Login:    "login_1",
					Password: password,
				}, nil)

				return repository
			},
			verify: func(repository userRepository) {
				Verify(repository, Once()).FindOneByLogin(
					AnyContext(),
					Exact("login_1"),
				)
			},
			wantToken: true,
		},
		{
			name: "user not found",
			repository: func() userRepository {
				repository := Mock[userRepository]()
				WhenDouble(repository.FindOneByLogin(
					AnyContext(),
					Exact("login_2"),
				)).ThenReturn(nil, nil)

				return repository
			},
			verify: func(repository userRepository) {
				Verify(repository, Once()).FindOneByLogin(
					AnyContext(),
					Exact("login_2"),
				)
			},
			wantErr: ErrInvalidCredentials,
		},
		{
			name: "invalid password",
			repository: func() userRepository {
				repository := Mock[userRepository]()
				password, _ := bcrypt.NewHash("password_invalid", 1)
				WhenDouble(repository.FindOneByLogin(
					AnyContext(),
					Exact("login_3"),
				)).ThenReturn(&entity.User{
					Login:    "login_3",
					Password: password,
				}, nil)

				return repository
			},
			verify: func(repository userRepository) {
				Verify(repository, Once()).FindOneByLogin(
					AnyContext(),
					Exact("login_3"),
				)
			},
			wantErr: ErrInvalidCredentials,
		},
		{
			name: "db error",
			repository: func() userRepository {
				repository := Mock[userRepository]()
				WhenDouble(repository.FindOneByLogin(
					AnyContext(),
					Exact("login_4"),
				)).ThenReturn(nil, someErr)

				return repository
			},
			verify: func(repository userRepository) {
				Verify(repository, Once()).FindOneByLogin(
					AnyContext(),
					Exact("login_4"),
				)
			},
			wantErr: someErr,
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetUp(t)
			i++
			repository := tt.repository()
			jwt := jwt.New(fmt.Sprintf("secret_%d", i))
			manager := NewUserManager(repository, jwt)

			token, err := manager.Authorize(context.Background(), fmt.Sprintf("login_%d", i), fmt.Sprintf("password_%d", i))
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			if tt.wantToken {
				assert.NotEmpty(t, token)
				claims, err := jwt.Decode(token)
				require.NoError(t, err)
				assert.Equal(t, fmt.Sprintf("login_%d", i), claims.Subject)
			} else {
				assert.Empty(t, token)
			}

			tt.verify(repository)
		})
	}
}
