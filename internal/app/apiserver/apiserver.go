package apiserver

import (
	"context"
	"effectiveMobile/internal/store/sqlstore"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"time"
)

func Start(config *Config) error {
	db, err := newDB(config.Postgres.URL)
	if err != nil {
		return err
	}
	defer db.Close()
	store := sqlstore.New(db)
	srv := newServer(store, config)
	srv.configureRouter()
	fmt.Println(config.Server.Port)
	return http.ListenAndServe(config.Server.Port, srv.router)
}

func newDB(dbURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse DATABASE_URL: %w", err)
	}
	cfg.MaxConns = 25
	cfg.MinConns = 5
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)

	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return pool, nil
}
