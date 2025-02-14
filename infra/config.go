package infra

import (
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	ServerName         string
	ServerPort         string
	Environment        string
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBDatabase         string
	DBSSLMode          string
	DBDriver           string
	SignatureToken     string
	AwsAccessKeyID     string
	AwsSecretAccessKey string
	AwsRegion          string
	GoogleMapsKey      string
	SignatureTokenSimp string
	RedisUrl           string
}

func NewConfig() Config {
	if os.Getenv("ENVIRONMENT") == "" {
		if err := godotenv.Load(".env"); err != nil {
			panic("Error loading env file")
		}
	}

	return Config{
		ServerName:         os.Getenv("SERVER_NAME"),
		ServerPort:         os.Getenv("SERVER_PORT"),
		Environment:        os.Getenv("ENVIRONMENT"),
		DBHost:             os.Getenv("DB_HOST"),
		DBPort:             os.Getenv("DB_PORT"),
		DBUser:             os.Getenv("DB_USER"),
		DBPassword:         os.Getenv("DB_PASSWORD"),
		DBDatabase:         os.Getenv("DB_DATABASE"),
		DBSSLMode:          os.Getenv("DB_SSL_MODE"),
		DBDriver:           os.Getenv("DB_DRIVER"),
		SignatureToken:     os.Getenv("SIGNATURE_STRING"),
		SignatureTokenSimp: os.Getenv("SIGNATURE_STRING_SIMP"),
		AwsAccessKeyID:     os.Getenv("AWS_ACCESS_KEY"),
		AwsSecretAccessKey: os.Getenv("AWS_SECRET_KEY"),
		AwsRegion:          os.Getenv("AWS_REGION"),
		GoogleMapsKey:      os.Getenv("GOOGLE_MAPS_KEY"),
		RedisUrl:           os.Getenv("REDIS_URL"),
	}
}
