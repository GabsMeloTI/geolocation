package login

type RequestLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

type ResponseLogin struct {
	Token string `json:"token"`
}

type RequestCreateUser struct {
	Email           string `json:"email"`
	Name            string `json:"name"`
	Telephone       string `json:"telephone"`
	Document        string `json:"document"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	TypePerson      int64  `json:"type_person"`
	Token           string `json:"token"`
}

type ResponseCreateUser struct {
	Token string `json:"token"`
}
