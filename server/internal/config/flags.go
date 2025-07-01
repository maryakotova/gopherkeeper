package config

import (
	"flag"
	"os"
)

type Flags struct {
	RunAddress  string
	DatabaseURI string
	JWTKey      string
}

func ParseFlags() *Flags {

	var flags Flags

	flag.StringVar(&flags.RunAddress, "a", "localhost:8080", "адрес и порт запуска сервера")
	flag.StringVar(&flags.DatabaseURI, "d", "host=localhost user=gopherkeeper password=test dbname=gopherkeeper sslmode=disable", "адрес подключения к базе данных")
	flag.StringVar(&flags.JWTKey, "j", "GopherKeeper_Secret_Key", "Ключ для JWT токена")

	if envRunAddress := os.Getenv("RUN_ADDRESS"); envRunAddress != "" {
		flags.RunAddress = envRunAddress
	}

	if envDatabaseURI := os.Getenv("DATABASE_URI"); envDatabaseURI != "" {
		flags.DatabaseURI = envDatabaseURI
	}

	if envJWTKey := os.Getenv("JWT_KEY"); envJWTKey != "" {
		flags.JWTKey = envJWTKey
	}

	return &flags
}
