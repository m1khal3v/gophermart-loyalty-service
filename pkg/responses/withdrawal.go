package responses

import (
	"time"
)

type Withdrawal struct {
	Order       uint64    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
