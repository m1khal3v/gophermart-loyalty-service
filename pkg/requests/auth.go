package requests

type Auth struct {
	Login    string `json:"login" valid:"required,stringlength(3|32),matches('^[0-9A-Za-z_-]+$')"`
	Password string `json:"password" valid:"required,stringlength(8|64)"`
}
