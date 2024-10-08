package responses

import (
	"time"
)

type Order struct {
	Number     uint64    `json:"number,string"`
	Status     string    `json:"status"`
	Accrual    *float64  `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}
