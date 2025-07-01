package postgres

import (
	"GophKeeper/server/internal/models"
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type PostgresStorage struct {
	db  *sql.DB
	mtx sync.RWMutex
}

func NewPostgresStorage(dbDSN string) (*PostgresStorage, error) {
	db, err := sql.Open("pgx", dbDSN)
	if err != nil {
		err = fmt.Errorf("не удалось подключиться к бд: %w", err)
		return nil, err
	}
	return &PostgresStorage{
		db: db,
	}, nil
}

func (s *PostgresStorage) Bootstrap(ctx context.Context) error {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
	CREATE TABLE IF NOT EXISTS users (
		user_id SERIAL PRIMARY KEY,
		user_name VARCHAR(50) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL
	);
	`

	_, err = tx.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	query = `
	CREATE TABLE IF NOT EXISTS files (
		id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id     INT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,	
		name        VARCHAR(255) NOT NULL,
		type        VARCHAR(20),
		meta        TEXT,
		created_at  TIMESTAMP NOT NULL,
		updated_at  TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_files_user_id ON files(user_id);
	`
	_, err = tx.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	query = `
	CREATE TABLE IF NOT EXISTS content (
		id      UUID PRIMARY KEY REFERENCES files(id) ON DELETE CASCADE,
		content BYTEA
	);
	`

	_, err = tx.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error creating tables: %v", err)
	}

	return nil
}

func (s *PostgresStorage) CreateUser(ctx context.Context, login string, hashedPassword string) (userID int, err error) {

	query := `
	INSERT INTO users (user_name, password)
		VALUES ($1, $2)
		RETURNING user_id;
	`

	s.mtx.Lock()
	err = s.db.QueryRowContext(ctx, query, login, hashedPassword).Scan(&userID)
	s.mtx.Unlock()
	if err != nil {
		return -1, err
	}

	log.Printf("[Repo]: создан пользователь с ID %v", userID)
	return
}

func (s *PostgresStorage) CLose() error {
	return s.db.Close()
}

func (s *PostgresStorage) GetUserAuthData(ctx context.Context, login string) (userID int, hashedPassword string, err error) {

	query := `
	SELECT user_id, password 
		FROM users 
		WHERE user_name = $1
	`
	s.mtx.Lock()
	err = s.db.QueryRowContext(ctx, query, login).Scan(&userID, &hashedPassword)
	s.mtx.Unlock()
	if err != nil {
		return -1, "", err
	}

	log.Printf("[Repo]: найден пользователь с ID %v", userID)
	return
}

func (s *PostgresStorage) InsertData(ctx context.Context, userID int, data []byte, fileName string, metadata string, createddAt time.Time) (id uuid.UUID, err error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	defer tx.Rollback()

	insertFileQuery := `
	INSERT INTO files (user_id, name, meta, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id;
	`

	insertContentQuery := `
	INSERT INTO content (id, content)
		VALUES ($1, $2);
	`

	s.mtx.Lock()
	err = tx.QueryRowContext(ctx, insertFileQuery, userID, fileName, metadata, createddAt, createddAt).Scan(&id)
	if err == nil {
		_, err = tx.ExecContext(ctx, insertContentQuery, id, data)
	}
	s.mtx.Unlock()

	if err != nil {
		return
	}

	if err = tx.Commit(); err != nil {
		err = fmt.Errorf("error inserting data: %w", err)
		return
	}

	return

}

func (s *PostgresStorage) UpdateData(ctx context.Context, fileID uuid.UUID, userID int, data []byte, fileName string, metadata string) error {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()
	updateFileQuery := `
		UPDATE files 
			SET name = $1, meta = $2, updated_at = $3
			WHERE id = $4
			AND user_id = $5;
	`

	updateContentQuery := `
		UPDATE content 
			SET content = $1
			WHERE id = $2;
	`

	s.mtx.Lock()
	_, err = tx.ExecContext(ctx, updateFileQuery, fileName, metadata, now, fileID, userID)
	if err == nil {
		_, err = tx.ExecContext(ctx, updateContentQuery, data, fileID)
	}
	s.mtx.Unlock()

	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		err = fmt.Errorf("error updating tables: %w", err)
		return err
	}

	return nil

}

func (s *PostgresStorage) DeleteData(ctx context.Context, fileID uuid.UUID, userID int) error {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	deleteQuery := `
		DELETE files
		WHERE id = $1
		AND user_id = $2;
	`

	s.mtx.Lock()
	_, err = tx.ExecContext(ctx, deleteQuery, fileID, userID)
	s.mtx.Unlock()

	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		err = fmt.Errorf("error deleting data: %w", err)
		return err
	}

	return nil
}

func (s *PostgresStorage) GetUpdatedSince(ctx context.Context, userID int, lastUpdate time.Time) (data []models.SyncData, err error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	defer tx.Rollback()

	selectQuery := `
		SELECT id, name, meta, content
			FROM files f
			INNER JOIN content c 
				ON f.id = c.id
			WHERE user_id = $1
			AND updated_at > $2;
	`
	s.mtx.Lock()
	rows, err := tx.QueryContext(ctx, selectQuery, userID, lastUpdate)
	s.mtx.Unlock()

	if err != nil {
		return
	}

	for rows.Next() {
		var row models.SyncData
		err = rows.Scan(&row.FileID, &row.FileName, &row.Metadata, row.Content)
		if err != nil {
			err = fmt.Errorf("ошибка при считывании строки: %w", err)
			return
		}
		data = append(data, row)
	}

	if err = tx.Commit(); err != nil {
		err = fmt.Errorf("error reading tables: %w", err)
		return
	}

	return
}

func (s *PostgresStorage) GetDataByFileID(ctx context.Context, fileID uuid.UUID, userID int) (data []byte, fileName string, metadata string, updatedAt time.Time, err error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	defer tx.Rollback()

	selectQuery := `
		SELECT name, meta, updated_at, content
			FROM files f
			INNER JOIN content c 
				ON f.id = c.id
			WHERE id = $1
			AND user_id = $2;
	`

	s.mtx.Lock()
	err = tx.QueryRowContext(ctx, selectQuery, fileID, userID).Scan(&fileName, &metadata, &updatedAt, &data)
	s.mtx.Unlock()

	if err != nil {
		return
	}

	if err = tx.Commit(); err != nil {
		err = fmt.Errorf("error reading tables: %w", err)
		return
	}

	return
}

func (s *PostgresStorage) GetUpdatedAtByID(ctx context.Context, fileID uuid.UUID, userID int) (updatedAt time.Time, err error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	defer tx.Rollback()

	selectFileQuery := `
		SELECT updated_at
			FROM files
			WHERE id = $1
			AND user_id = $2;
	`

	s.mtx.Lock()
	err = tx.QueryRowContext(ctx, selectFileQuery, fileID, userID).Scan(&updatedAt)
	s.mtx.Unlock()

	if err != nil {
		return
	}

	if err = tx.Commit(); err != nil {
		err = fmt.Errorf("error reading tables: %w", err)
		return
	}

	return
}
