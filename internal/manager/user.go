package manager

import (
	"context"
	"errors"

	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/jwt"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/gorm/types/bcrypt"
	"gorm.io/gorm"
)

var ErrLoginAlreadyExists = errors.New("login already exists")
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrUserNotFound = errors.New("user not found")

type userRepository interface {
	Create(ctx context.Context, entity *entity.User) error
	FindOneByLogin(ctx context.Context, login string) (*entity.User, error)
	FindByID(ctx context.Context, id uint32) (*entity.User, error)
}

type UserManager struct {
	userRepository userRepository
	jwt            *jwt.Container
}

func NewUserManager(userRepository userRepository, jwt *jwt.Container) *UserManager {
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

func (manager *UserManager) FindByID(ctx context.Context, id uint32) (*entity.User, error) {
	user, err := manager.userRepository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}
