package responses

type RegisterUser struct {
	ID    uint32 `json:"id"`
	Login string `json:"login"`
}
