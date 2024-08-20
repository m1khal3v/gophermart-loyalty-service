package requests

type Withdraw struct {
	Order uint64  `json:"order,string" valid:"required,luhn"`
	Sum   float64 `json:"sum" valid:"required,positive"`
}
