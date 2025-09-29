package database

type Config struct {
	Host        string
	Port        string
	User        string
	Password    string
	Database    string
	SSLMode     string
	Driver      string
	Environment string
	DatabaseSP  string
}
