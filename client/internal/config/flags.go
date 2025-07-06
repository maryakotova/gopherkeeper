package config

import (
	"flag"
	"os"
)

type Flags struct {
	RunAddress  string
	DatabaseURI string
	TokenPath   string
}

func ParseFlags() *Flags {

	var flags Flags

	flag.StringVar(&flags.RunAddress, "a", "localhost:8080", "адрес и порт запуска сервиса")
	flag.StringVar(&flags.DatabaseURI, "d", "host=localhost user=gopherkeeperclient password=test dbname=gopherkeeperclient sslmode=disable", "адрес подключения к базе данных")
	flag.StringVar(&flags.TokenPath, "t", "./token.txt", "Путь к файла для хранения токена")

	if envRunAddress := os.Getenv("RUN_ADDRESS"); envRunAddress != "" {
		flags.RunAddress = envRunAddress
	}

	if envDatabaseURI := os.Getenv("DATABASE_URI"); envDatabaseURI != "" {
		flags.DatabaseURI = envDatabaseURI
	}

	if envTokenPath := os.Getenv("TOKEN_PATH"); envTokenPath != "" {
		flags.TokenPath = envTokenPath
	}

	return &flags
}
