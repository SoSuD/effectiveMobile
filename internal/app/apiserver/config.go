package apiserver

import (
	"github.com/joho/godotenv"
	"os"
)

type Server struct {
	Port string
}

type Postgres struct {
	URL string
}

type Zap struct {
	Level string
}

type ExternalService struct {
	AgifyURL       string
	GenderizeURL   string
	NationalizeURL string
}

type Config struct {
	Server          Server
	Postgres        Postgres
	Zap             Zap
	ExternalService ExternalService
}

func NewConfig() *Config {
	if err := godotenv.Load(); err != nil {
		return nil
	}

	return &Config{
		Server: Server{
			Port: os.Getenv("SERVER_PORT"),
		},
		Postgres: Postgres{
			URL: os.Getenv("DATABASE_URL"),
		},
		Zap: Zap{
			Level: os.Getenv("ZAP_LEVEL"),
		},
		ExternalService: ExternalService{
			AgifyURL:       os.Getenv("AGIFY_URL"),
			GenderizeURL:   os.Getenv("GENDERIZE_URL"),
			NationalizeURL: os.Getenv("NATIONALIZE_URL"),
		},
	}
}
