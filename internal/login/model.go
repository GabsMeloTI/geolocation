package login

type RequestLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

type ResponseLogin struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Token string `json:"token"`
}

type RequestCreateUser struct {
	Email           string `json:"email" validate:"required"`
	Name            string `json:"name" validate:"required"`
	Telephone       string `json:"telephone" validate:"required"`
	Document        string `json:"document" validate:"required"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	TypePerson      int64  `json:"type_person" validate:"required"`
	Token           string `json:"token"`
}

type ResponseCreateUser struct {
	Token string `json:"token"`
}
