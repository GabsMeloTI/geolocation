package email

type SmtpConfig struct {
	Email    string
	Password string
	Host     string
	Port     string
}

type EmailPlaceHolder struct {
	NameProvider     string
	EmailProvider    string
	PasswordProvider string
	AccessKey        string
	Link             string
}
