package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type Postgres struct {
	cfg *Config

	Conn *sql.DB
}

func New(cfg *Config) (*Postgres, error) {
	conn, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("connect database error: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping database error: %w", err)
	}

	return &Postgres{
		cfg:  cfg,
		Conn: conn,
	}, nil
}

func (p *Postgres) MigrateUP() error {
	m, err := migrate.New(p.cfg.MigrationURL, p.cfg.DSN)
	if err != nil {
		return fmt.Errorf("migrate create error: %w", err)
	}

	if err := m.Up(); err != nil {
		return fmt.Errorf("migrations up error: %w", err)
	}

	return nil
}

func (p *Postgres) Close() {
	if p.Conn != nil {
		p.Conn.Close()
	}
}
