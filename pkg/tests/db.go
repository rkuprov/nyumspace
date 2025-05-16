package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rkuprov/nyumspace/pkg/config"
)

func DBForTest(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	cfg, err := config.NewConfig()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	dbStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.PG.User,
		cfg.PG.Password,
		cfg.PG.Host,
		cfg.PG.Port,
		cfg.PG.DbName)
	db, err := pgxpool.New(ctx, dbStr)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	dbName := fmt.Sprintf("test%d", time.Now().UnixNano())
	_, err = db.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s OWNER %s", dbName, cfg.PG.User))
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	_, err = db.Exec(ctx, fmt.Sprintf("GRANT ALL PRIVILEGES ON DATABASE %s TO %s", dbName, cfg.PG.User))
	if err != nil {
		t.Fatalf("failed to grant privileges on test database: %v", err)
	}

	dbStr = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.PG.User,
		cfg.PG.Password,
		cfg.PG.Host,
		cfg.PG.Port,
		dbName)

	db, err = pgxpool.New(ctx, dbStr)
	if err != nil {
		db.Close()
		t.Fatalf("failed to connect to test database: %v", err)
	}

	runMigrations(ctx, db)

	return db
}

func RemoveDBForTest(p *pgxpool.Pool) error {
	cfg, _ := config.NewConfig()
	dbNameToDelete := p.Config().ConnConfig.Database

	pool, err := pgxpool.New(context.Background(), fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.PG.User,
		cfg.PG.Password,
		cfg.PG.Host,
		cfg.PG.Port,
		cfg.PG.DbName,
	))
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	_, err = pool.Exec(context.Background(), fmt.Sprintf(`drop database %s`, dbNameToDelete))
	if err != nil {
		return err
	}

	return nil
}

func runMigrations(ctx context.Context, db *pgxpool.Pool) {
}
