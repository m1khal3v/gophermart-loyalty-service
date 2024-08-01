package manager

import (
	"errors"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/entity"
	"github.com/m1khal3v/gophermart-loyalty-service/internal/repository"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/gorm/types/bcrypt"
	"gorm.io/gorm"
)

var ErrLoginAlreadyExists = errors.New("login already exists")

type UserManager struct {
	userRepository *repository.UserRepository
}

func NewUserManager(userRepository *repository.UserRepository) *UserManager {
	return &UserManager{
		userRepository: userRepository,
	}
}

func (manager *UserManager) RegisterUser(login, password string) (*entity.User, error) {
	hash, err := bcrypt.NewHash(password, bcrypt.RecommendedCost)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		Login:    login,
		Password: hash,
	}

	if err := manager.userRepository.Create(user); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrLoginAlreadyExists
		}

		return nil, err
	}

	return user, nil
}
