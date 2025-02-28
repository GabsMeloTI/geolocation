package login

type RequestLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

type ResponseLogin struct {
	Token string `json:"token"`
}
