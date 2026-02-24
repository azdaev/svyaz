package repo

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

type Repo struct {
	db *sql.DB
}

func New(dbPath string, migrationsDir string) (*Repo, error) {
	if err := os.MkdirAll(".", 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, fmt.Errorf("set dialect: %w", err)
	}

	if err := goose.Up(db, migrationsDir); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return &Repo{db: db}, nil
}

func (r *Repo) Close() error {
	return r.db.Close()
}
