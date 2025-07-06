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

func (s *PostgresStorage) InsertFile(ctx context.Context, data []byte, fileName string, metadata string, id_server uuid.UUID, created_at time.Time) (id uuid.UUID, err error) {

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
	err = tx.QueryRowContext(ctx, insertFileQuery, id_server, fileName, metadata, data, created_at, created_at).Scan(&id)
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

// func (s *PostgresStorage) UpdateIDServer(ctx context.Context, id uuid.UUID, id_server uuid.UUID) error {

// 	tx, err := s.db.BeginTx(ctx, nil)
// 	if err != nil {
// 		return err
// 	}
// 	defer tx.Rollback()

// 	updateFileQuery := `
// 	UPDATE files
// 			SET id_server = $1
// 			WHERE id = $2;
// 	`
// 	s.mtx.Lock()
// 	_, err = tx.ExecContext(ctx, updateFileQuery, id, id_server)
// 	s.mtx.Unlock()

// 	if err != nil {
// 		return err
// 	}

// 	if err = tx.Commit(); err != nil {
// 		err = fmt.Errorf("error updating tables: %w", err)
// 		return err
// 	}

// 	return nil
// }

func (s *PostgresStorage) UpdateFile(ctx context.Context, id uuid.UUID, data []byte, fileName string, metadata string, updated_at time.Time) error {

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
	_, err = tx.ExecContext(ctx, updateFileQuery, fileName, metadata, data, updated_at, id)
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
			FROM key_date 
			WHERE id = $1;
	`
	s.mtx.Lock()
	err := s.db.QueryRowContext(ctx, query, lastSyncKey).Scan(&timestamp)
	s.mtx.Unlock()

	if err != nil {
		return timestamp
	}
	return timestamp

}

func (s *PostgresStorage) UpdateLastSync(ctx context.Context, lastSymc time.Time) time.Time {

	var timestamp time.Time

	query := `
		INSERT INTO key_timestamps (key, value)
			VALUES ($1, $2)
			ON CONFLICT (key)
			DO UPDATE SET
				value = EXCLUDED.value;
	`
	s.mtx.Lock()
	_, err := s.db.ExecContext(ctx, query, lastSyncKey, lastSymc)
	s.mtx.Unlock()

	if err != nil {
		return timestamp
	}
	return timestamp

}
