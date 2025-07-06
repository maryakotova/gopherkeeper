package storage

import (
	"GophKeeper/server/internal/models"
	"GophKeeper/server/internal/storage/postgres"
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	CreateUser(ctx context.Context, login string, hashedPassword string) (userID int, err error)
	GetUserAuthData(ctx context.Context, login string) (userID int, hashedPassword string, err error)
	Close() error
	Bootstrap(ctx context.Context) error
	InsertData(ctx context.Context, userID int, data []byte, fileName string, metadata string, createddAt time.Time) (id uuid.UUID, err error)
	DeleteData(ctx context.Context, fileID uuid.UUID, userID int) error
	UpdateData(ctx context.Context, fileID uuid.UUID, userID int, data []byte, fileName string, metadata string) error
	GetDataByFileID(ctx context.Context, fileID uuid.UUID, userID int) (data []byte, fileName string, metadata string, updatedAt time.Time, err error)
	GetUpdatedSince(ctx context.Context, userID int, lastUpdate time.Time) (data []models.SyncData, err error)
	GetUpdatedAtByID(ctx context.Context, fileID uuid.UUID, userID int) (updatedAt time.Time, err error)
}

type StorageFactory struct{}

func (sf *StorageFactory) NewStorage(dbDSN string) (Repository, error) {
	postgres, err := postgres.NewPostgresStorage(dbDSN)
	if err != nil {
		return nil, err
	}

	err = postgres.Bootstrap(context.TODO())

	return postgres, err
}
