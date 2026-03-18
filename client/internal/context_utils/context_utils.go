package contextutils

import (
	"context"
	"fmt"
	"os"
)

type contextKey string

const (
	ServerAddrKey  contextKey = "server_addr"
	DatabaseURIKey contextKey = "db_uri"
	TokenPathKey   contextKey = "token_path"
)

func WithDatabaseURL(ctx context.Context, dbURI string) context.Context {
	return context.WithValue(ctx, DatabaseURIKey, dbURI)
}

func WithTokenPath(ctx context.Context, path string) context.Context {
	return context.WithValue(ctx, TokenPathKey, path)
}

func WithServerAddr(ctx context.Context, addr string) context.Context {
	return context.WithValue(ctx, ServerAddrKey, addr)
}

func GetDatabaseURL(ctx context.Context) (value string, err error) {
	value, _ = ctx.Value(DatabaseURIKey).(string)
	if value == "" {
		err = fmt.Errorf("адрес подключения к базе данных не найден")
	}
	return
}

func GetTokenPath(ctx context.Context) (value string, err error) {
	value, _ = ctx.Value(TokenPathKey).(string)
	if value == "" {
		err = fmt.Errorf("путь к файла для хранения токена не найден")
	}
	return
}

func GetServerAddr(ctx context.Context) (value string, err error) {
	value, _ = ctx.Value(ServerAddrKey).(string)
	if value == "" {
		err = fmt.Errorf("адрес подключения к базе данных не найден")
	}
	return
}

func SaveTokenToFile(ctx context.Context, token string) error {
	path, err := GetTokenPath(ctx)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(token), 0600)
}

func LoadTokenFromFile(ctx context.Context) (string, error) {

	path, err := GetTokenPath(ctx)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
