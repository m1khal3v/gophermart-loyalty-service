package entity

import (
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/gorm/types/bcrypt"
	"time"
)

type User struct {
	ID        uint32      `gorm:"primaryKey;autoIncrement"`
	Login     string      `gorm:"not null;uniqueIndex:idx_login"`
	Password  bcrypt.Hash `gorm:"not null"`
	CreatedAt time.Time
}
