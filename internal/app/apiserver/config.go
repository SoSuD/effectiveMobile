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

type Config struct {
	Server   Server
	Postgres Postgres
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
	}
}
