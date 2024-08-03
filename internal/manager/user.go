package manager

import (
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

func (manager *UserManager) RegisterUser(login, password string) (string, error) {
	hash, err := bcrypt.NewHash(password, bcrypt.RecommendedCost)
	if err != nil {
		return "", err
	}

	user := &entity.User{
		Login:    login,
		Password: hash,
	}

	if err := manager.userRepository.Create(user); err != nil {
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

func (manager *UserManager) AuthUser(login, password string) (string, error) {
	user, err := manager.userRepository.FindOneByLogin(login)
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
