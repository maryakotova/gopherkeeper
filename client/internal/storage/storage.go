package storage

import (
	"GophKeeper/client/internal/storage/postgres"
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	Close() error
	Bootstrap(ctx context.Context) error
	InsertFile(ctx context.Context, data []byte, fileName string, metadata string, id_server uuid.UUID, createdAt time.Time) (id uuid.UUID, err error)
	UpdateFile(ctx context.Context, id uuid.UUID, data []byte, fileName string, metadata string, updatedAt time.Time) error
	GetServerIDByFileID(ctx context.Context, id uuid.UUID) (server_id uuid.UUID, err error)
	GetLastSync(ctx context.Context) time.Time
	UpdateLastSync(ctx context.Context, lastSync time.Time) error
	SaveOrUpdate(ctx context.Context, id_server uuid.UUID, fileName string, content []byte, meta string, createdAt time.Time, updatedAt time.Time) error
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
