package service

import (
	"GophKeeper/server/internal/models"
	"GophKeeper/server/internal/storage"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ControllerService interface {
	Download(ctx context.Context, fileID uuid.UUID, userID int) (data []byte, fileName string, metadata string, updatedAt time.Time, err error)
	Upload(ctx context.Context, userID int, data []byte, fileName string, metadata string, createdAt time.Time) (id uuid.UUID, err error)
	Update(ctx context.Context, id uuid.UUID, userID int, data []byte, fileName string, metadata string, updatedAt time.Time) (err error)
	GetDataForSync(ctx context.Context, userID int, lastUpdate time.Time) (data []models.SyncData, err error)
}

type controllerService struct {
	repo storage.Repository
}

func NewControllerService(repo storage.Repository) ControllerService {
	return &controllerService{
		repo: repo,
	}
}

func (s controllerService) Download(ctx context.Context, fileID uuid.UUID, userID int) (data []byte, fileName string, metadata string, updatedAt time.Time, err error) {

	data, fileName, metadata, updatedAt, err = s.repo.GetDataByFileID(ctx, fileID, userID)
	return
}

func (s controllerService) Upload(ctx context.Context, userID int, data []byte, fileName string, metadata string, createddAt time.Time) (id uuid.UUID, err error) {

	id, err = s.repo.InsertData(ctx, userID, data, fileName, metadata, createddAt)
	if err != nil {
		return
	}

	return

}

func (s controllerService) Update(ctx context.Context, id uuid.UUID, userID int, data []byte, fileName string, metadata string, updatedAt time.Time) (err error) {

	dbUpdatedAt, err := s.repo.GetUpdatedAtByID(ctx, id, userID)
	if err != nil {
		return
	}

	if dbUpdatedAt.After(updatedAt) || dbUpdatedAt.Equal(updatedAt) {
		return fmt.Errorf("конфликт версий: версия данных в хранилище актуальнее полученной")
	}

	return s.repo.UpdateData(ctx, id, userID, data, fileName, metadata)

}

func (s controllerService) GetDataForSync(ctx context.Context, userID int, lastUpdate time.Time) (data []models.SyncData, err error) {

	return s.repo.GetUpdatedSince(ctx, userID, lastUpdate)

}
