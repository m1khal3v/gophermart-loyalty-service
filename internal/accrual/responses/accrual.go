package responses

const (
	AccrualStatusRegistered string = "REGISTERED"
	AccrualStatusProcessing string = "PROCESSING"
	AccrualStatusInvalid    string = "INVALID"
	AccrualStatusProcessed  string = "PROCESSED"
)

type Accrual struct {
	OrderID uint64   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual"`
}
