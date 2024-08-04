package entity

import (
	"github.com/m1khal3v/gophermart-loyalty-service/pkg/gorm/types/money"
	"time"
)

type Withdrawal struct {
	OrderID uint64 `gorm:"primaryKey;autoIncrement:false"`
	UserID  uint32 `gorm:"not null"`

	Sum money.Amount `gorm:"not null"`

	CreatedAt time.Time `gorm:"not null;autoCreateTime;index:idx_created_at,sort:desc"`
}
