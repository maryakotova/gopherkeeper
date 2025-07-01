package service

import (
	"GophKeeper/server/internal/auth"
	"GophKeeper/server/internal/customererrors"
	"GophKeeper/server/internal/storage"
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgconn"
)

const (
	duplicateKeyCode = "23505"
)

type AuthService interface {
	Register(ctx context.Context, login string, password string) error
	Login(ctx context.Context, login string, password string) (jwt string, err error)
}

type authService struct {
	repo storage.Repository
}

func NewAuthService(repo storage.Repository) AuthService {
	return &authService{
		repo: repo,
	}
}

func (s authService) Register(ctx context.Context, login string, password string) error {

	hash, err := auth.HashPassword(password)
	if err != nil {
		err = fmt.Errorf("[AuthService]: ошибка при хэшировании пароля: %w", err)
		return err
	}

	_, err = s.repo.CreateUser(ctx, login, hash)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == duplicateKeyCode {
				log.Printf("[AuthService]: пользователь с таким именем уже существует")
				return customererrors.ErrUsernameTaken
			}
		}
		err = fmt.Errorf("[AuthService]: ошибка при создании пользователя: %w", err)
		return err
	}

	return nil
}

func (s authService) Login(ctx context.Context, login string, password string) (jwt string, err error) {
	// hash, err := auth.HashPassword(password)
	// if err != nil {
	// 	err = fmt.Errorf("[AuthService]: ошибка при хэшировании пароля: %w", err)
	// 	log.Print(err)
	// 	return "", err
	// }

	userID, dbPassword, err := s.repo.GetUserAuthData(ctx, login)
	if err != nil {
		err = fmt.Errorf("[AuthService]: пользователь не найден: %w", err)
		log.Print(err)
		return "", err
	}

	err = auth.СheckPassword(dbPassword, password)
	if err != nil {
		err = fmt.Errorf("[AuthService]: неверный пароль: %w", err)
		log.Print(err)
		return "", err
	}

	jwt, err = auth.GenerateJWT(userID)
	if err != nil {
		err = fmt.Errorf("[AuthService]: ошибка при генерации токена: %w", err)
		log.Print(err)
		return "", err
	}

	return
}
