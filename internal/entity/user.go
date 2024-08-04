package entity

import (
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/gorm/types/bcrypt"
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/gorm/types/money"
	"time"
)

type User struct {
	ID uint32 `gorm:"primaryKey;autoIncrement"`

	Login    string      `gorm:"not null;size:32;uniqueIndex:idx_login"`
	Password bcrypt.Hash `gorm:"not null"`

	Balance   money.Amount `gorm:"not null;default:0"`
	Withdrawn money.Amount `gorm:"not null;default:0"`

	Orders      []Order      `gorm:"foreignKey:UserID"`
	Withdrawals []Withdrawal `gorm:"foreignKey:UserID"`

	CreatedAt time.Time `gorm:"not null;autoCreateTime"`
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime"`
}
