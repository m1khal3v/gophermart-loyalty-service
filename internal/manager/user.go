package manager

import (
	"context"
	"errors"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/jwt"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/repository"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/gorm/types/bcrypt"
	"gorm.io/gorm"
)

var ErrLoginAlreadyExists = errors.New("login already exists")
var ErrInvalidCredentials = errors.New("invalid credentials")

type UserManager struct {
	userRepository *repository.UserRepository
	jwt            *jwt.Container
}

func NewUserManager(userRepository *repository.UserRepository, jwt *jwt.Container) *UserManager {
	return &UserManager{
		userRepository: userRepository,
		jwt:            jwt,
	}
}

func (manager *UserManager) Register(ctx context.Context, login, password string) (string, error) {
	hash, err := bcrypt.NewHash(password, bcrypt.RecommendedCost)
	if err != nil {
		return "", err
	}

	user := &entity.User{
		Login:    login,
		Password: hash,
	}

	if err := manager.userRepository.Create(ctx, user); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return "", ErrLoginAlreadyExists
		}

		return "", err
	}

	token, err := manager.jwt.Encode(user.ID, user.Login)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (manager *UserManager) Authorize(ctx context.Context, login, password string) (string, error) {
	user, err := manager.userRepository.FindOneByLogin(ctx, login)
	if err != nil {
		return "", err
	}

	if user == nil {
		return "", ErrInvalidCredentials
	}
	if err := user.Password.CompareWithPassword(password); err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := manager.jwt.Encode(user.ID, user.Login)
	if err != nil {
		return "", err
	}

	return token, nil
}
