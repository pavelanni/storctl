package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/pavelanni/storctl/internal/config"
	"github.com/pavelanni/storctl/internal/logger"
	"github.com/pavelanni/storctl/internal/types"
)

type Storage struct {
	db     *sql.DB
	logger *slog.Logger
}

func New(cfg *config.Config) (*Storage, error) {
	connStr := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		cfg.Storage.Postgres.Host,
		cfg.Storage.Postgres.Port,
		cfg.Storage.Postgres.Database,
		cfg.Storage.Postgres.User,
		cfg.Storage.Postgres.Password)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres db: %w", err)
	}

	return &Storage{db: db, logger: logger.Get()}, nil
}

func (s *Storage) Save(lab *types.Lab) error {
	metadata, _ := json.Marshal(lab.ObjectMeta)
	spec, _ := json.Marshal(lab.Spec)
	status, _ := json.Marshal(lab.Status)

	_, err := s.db.Exec(`
        INSERT INTO labs (name, metadata, spec, status)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (name) DO UPDATE
        SET metadata = $2, spec = $3, status = $4, updated_at = NOW()
    `, lab.Name, metadata, spec, status)

	if err != nil {
		return fmt.Errorf("failed to save lab: %w", err)
	}
	return nil
}

func (s *Storage) Get(name string) (*types.Lab, error) {
	var metadata, spec, status []byte

	err := s.db.QueryRow(`
        SELECT metadata, spec, status
        FROM labs
        WHERE name = $1 AND deleted_at IS NULL
    `, name).Scan(&metadata, &spec, &status)

	if err != nil {
		return nil, fmt.Errorf("failed to get lab: %w", err)
	}

	lab := &types.Lab{
		TypeMeta: types.TypeMeta{
			APIVersion: "v1",
			Kind:       "Lab",
		},
	}
	err = json.Unmarshal(metadata, &lab.ObjectMeta)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}
	err = json.Unmarshal(spec, &lab.Spec)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal spec: %w", err)
	}
	err = json.Unmarshal(status, &lab.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal status: %w", err)
	}
	lab.Name = name

	return lab, nil
}

func (s *Storage) List(showDeleted bool) ([]*types.Lab, error) {
	s.logger.Debug("Executing List query", "showDeleted", showDeleted)

	query := `SELECT name, metadata, spec, status FROM labs`
	if !showDeleted {
		query += ` WHERE deleted_at IS NULL`
	}

	rows, err := s.db.Query(query)
	if err != nil {
		s.logger.Error("Query failed", "error", err)
		return nil, fmt.Errorf("failed to query labs: %w", err)
	}
	s.logger.Debug("Query executed successfully")
	defer func() {
		if err := rows.Close(); err != nil {
			s.logger.Error("failed to close rows", "error", err)
		}
	}()

	var labs []*types.Lab
	rowCount := 0
	for rows.Next() {
		rowCount++
		var name string
		var metadata, spec, status []byte
		err = rows.Scan(&name, &metadata, &spec, &status)
		if err != nil {
			s.logger.Error("Failed to scan row", "error", err, "rowCount", rowCount)
			return nil, fmt.Errorf("failed to scan lab: %w", err)
		}
		s.logger.Debug("Scanned row", "name", name, "rowCount", rowCount)
		lab := &types.Lab{
			TypeMeta: types.TypeMeta{
				APIVersion: "v1",
				Kind:       "Lab",
			},
		}
		err = json.Unmarshal(metadata, &lab.ObjectMeta)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		err = json.Unmarshal(spec, &lab.Spec)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal spec: %w", err)
		}
		err = json.Unmarshal(status, &lab.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal status: %w", err)
		}
		lab.Name = name
		labs = append(labs, lab)
	}

	// Check for errors from iterating over rows
	if err = rows.Err(); err != nil {
		s.logger.Error("Error iterating over rows", "error", err)
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	s.logger.Debug("List completed", "totalRows", rowCount, "labsReturned", len(labs))
	return labs, nil
}

func (s *Storage) Delete(name string) error {
	_, err := s.db.Exec(`
		UPDATE labs SET deleted_at = NOW() WHERE name = $1
	`, name)
	if err != nil {
		return fmt.Errorf("failed to delete lab: %w", err)
	}
	return nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}
