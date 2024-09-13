package entity

import (
	"time"

	"github.com/m1khal3v/gophermart-loyalty-service/pkg/gorm/types/money"
)

const (
	OrderStatusNew        string = "NEW"
	OrderStatusProcessing string = "PROCESSING"
	OrderStatusInvalid    string = "INVALID"
	OrderStatusProcessed  string = "PROCESSED"
)

type Order struct {
	ID     uint64 `gorm:"primaryKey;autoIncrement:false"`
	UserID uint32 `gorm:"not null"`

	Status  string       `gorm:"not null;size:16;default:'NEW'"`
	Accrual money.Amount `gorm:"not null;default:0"`

	CreatedAt time.Time `gorm:"not null;autoCreateTime;index:idx_order_created_at,sort:desc"`
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime"`
}
