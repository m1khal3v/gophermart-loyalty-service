package entity

import (
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/gorm/types/bcrypt"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/gorm/types/money"
	"time"
)

type User struct {
	ID       uint32      `gorm:"primaryKey;autoIncrement'"`
	Login    string      `gorm:"not null;uniqueIndex:idx_login"`
	Password bcrypt.Hash `gorm:"not null"`

	Balance   money.Amount `gorm:"not null;default:0"`
	Withdrawn money.Amount `gorm:"not null;default:0"`

	CreatedAt time.Time
}
