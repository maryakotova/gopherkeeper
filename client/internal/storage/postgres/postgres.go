package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	lastSyncKey = "last_sync"
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
	CREATE TABLE IF NOT EXISTS files (
		id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		id_server   UUID,
		name        VARCHAR(255) NOT NULL,
		type        VARCHAR(20),
		meta        TEXT,
		content 	BYTEA,
		created_at  TIMESTAMP NOT NULL,
		updated_at  TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS files_id_server ON files(id_server);
	`
	_, err = tx.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	querySync := `
	CREATE TABLE IF NOT EXISTS key_timestamps (
		key TEXT PRIMARY KEY,
		value TIMESTAMP NOT NULL
	);
	`
	_, err = tx.ExecContext(ctx, querySync)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error creating tables: %v", err)
	}

	return nil
}

func (s *PostgresStorage) Close() error {
	return s.db.Close()
}

func (s *PostgresStorage) InsertFile(ctx context.Context, data []byte, fileName string, metadata string, id_server uuid.UUID, createdAt time.Time) (id uuid.UUID, err error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	defer tx.Rollback()

	insertFileQuery := `
	INSERT INTO files (id_server, name, meta, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id;
	`

	s.mtx.Lock()
	err = tx.QueryRowContext(ctx, insertFileQuery, id_server, fileName, metadata, data, createdAt, createdAt).Scan(&id)
	s.mtx.Unlock()

	if err != nil {
		return
	}

	if err = tx.Commit(); err != nil {
		err = fmt.Errorf("error updating tables: %w", err)
		return
	}

	return
}

func (s *PostgresStorage) UpdateFile(ctx context.Context, id uuid.UUID, data []byte, fileName string, metadata string, updatedAt time.Time) error {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	updateFileQuery := `
		UPDATE files 
			SET name = $1, meta = $2, content = $3, updated_at = $4
			WHERE id = $5;
	`
	s.mtx.Lock()
	_, err = tx.ExecContext(ctx, updateFileQuery, fileName, metadata, data, updatedAt, id)
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

func (s *PostgresStorage) GetServerIDByFileID(ctx context.Context, id uuid.UUID) (server_id uuid.UUID, err error) {
	query := `
		SELECT id_server 
			FROM files 
			WHERE id = $1;
	`
	s.mtx.Lock()
	err = s.db.QueryRowContext(ctx, query, id).Scan(&server_id)
	s.mtx.Unlock()

	if err != nil {
		return
	}
	return
}

func (s *PostgresStorage) GetLastSync(ctx context.Context) time.Time {

	var timestamp time.Time

	query := `
		SELECT value 
			FROM key_timestamps 
			WHERE key = $1;
	`
	s.mtx.Lock()
	err := s.db.QueryRowContext(ctx, query, lastSyncKey).Scan(&timestamp)
	s.mtx.Unlock()

	if err != nil {
		return timestamp
	}
	return timestamp

}

func (s *PostgresStorage) UpdateLastSync(ctx context.Context, lastSync time.Time) error {

	query := `
		INSERT INTO key_timestamps (key, value)
			VALUES ($1, $2)
			ON CONFLICT (key)
			DO UPDATE SET
				value = EXCLUDED.value;
	`
	s.mtx.Lock()
	_, err := s.db.ExecContext(ctx, query, lastSyncKey, lastSync)
	s.mtx.Unlock()

	if err != nil {
		return fmt.Errorf("не удалось обновить дату последней синхронизации: %w", err)
	}
	return nil

}

func (s *PostgresStorage) SaveOrUpdate(ctx context.Context, id_server uuid.UUID, fileName string, content []byte, meta string, createdAt time.Time, updatedAt time.Time) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("начало транзакции: %w", err)
	}
	defer tx.Rollback()

	var id uuid.UUID

	queryExists := `
		SELECT id 
			FROM files 
			WHERE id_server = $1; 
	`

	err = tx.QueryRowContext(ctx, queryExists, id_server).Scan(&id)

	if err == nil && id != uuid.Nil {
		queryUpdate := `
			UPDATE files 
				SET name = $1, meta = $2, content = $3, created_at = $4, updated_at = $5
				WHERE id = $6;
		`
		_, err = tx.ExecContext(ctx, queryUpdate, fileName, meta, content, createdAt, updatedAt, id)
		if err != nil {
			return fmt.Errorf("ошибка при обновлении файла: %w", err)
		}
	} else if err == sql.ErrNoRows {
		queryInsert := `
			INSERT INTO files (id_server, name, meta, content, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6);
		`
		_, err = tx.ExecContext(ctx, queryInsert, id_server, fileName, meta, content, createdAt, updatedAt)
		if err != nil {
			return fmt.Errorf("ошибка при вставке файла: %w", err)
		}
	} else {
		return fmt.Errorf("ошибка выборки: %w", err)
	}

	return tx.Commit()
}
