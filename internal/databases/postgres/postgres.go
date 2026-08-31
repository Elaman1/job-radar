package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"job-radar/internal/config"
	"time"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(ctx context.Context, conf config.Postgres) (*Postgres, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		conf.Host, conf.Port, conf.User, conf.Password, conf.Database, conf.SSLMode,
	)

	parseConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse postgres connection string: %w", err)
	}

	parseConfig.MaxConns = conf.MaxConns
	parseConfig.MinConns = conf.MinConns
	parseConfig.MaxConnLifetime = time.Duration(conf.MaxConnLifeTime)
	parseConfig.MaxConnIdleTime = time.Duration(conf.MaxConnIdleTime)

	pool, err := pgxpool.NewWithConfig(ctx, parseConfig)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	if errPing := pool.Ping(ctx); errPing != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", errPing)
	}

	return &Postgres{pool: pool}, nil
}

func (p *Postgres) Pool() *pgxpool.Pool {
	return p.pool
}

func (p *Postgres) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}

func (p *Postgres) Close() {
	p.pool.Close()
}

func (p *Postgres) Begin(ctx context.Context) (pgx.Tx, error) {
	return p.pool.Begin(ctx)
}
