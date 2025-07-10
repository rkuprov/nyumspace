package tests

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"

	"github.com/rkuprov/nyumspace/migrations"
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
	masterDB, err := pgxpool.New(ctx, dbStr)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer masterDB.Close()

	dbName := fmt.Sprintf("test%d", time.Now().UnixNano())
	_, err = masterDB.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s OWNER %s", dbName, cfg.PG.User))
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	_, err = masterDB.Exec(ctx, fmt.Sprintf("GRANT ALL PRIVILEGES ON DATABASE %s TO %s", dbName, cfg.PG.User))
	if err != nil {
		t.Fatalf("failed to grant privileges on test database: %v", err)
	}

	dbStr = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.PG.User,
		cfg.PG.Password,
		cfg.PG.Host,
		cfg.PG.Port,
		dbName)

	db, err := pgxpool.New(ctx, dbStr)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	err = runMigrations(ctx, db)
	if err != nil {
		db.Close()
		t.Fatalf("failed to run migrations: %v", err)
	}

	return db
}

func CleanupTestDB(t *testing.T, db *pgxpool.Pool) {
	if db == nil {
		t.Fatal("db is nil")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbName := db.Config().ConnConfig.Database
	err := removeDBForTest(ctx, dbName)
	assert.NoError(t, err, "failed to remove test database")
}

func removeDBForTest(ctx context.Context, name string) error {
	cfg, err := config.NewConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	pool, err := pgxpool.New(ctx, fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
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

	// Terminate all connections to the database before dropping it
	_, err = pool.Exec(ctx,
		`SELECT pg_terminate_backend(pg_stat_activity.pid)
		FROM pg_stat_activity
		WHERE pg_stat_activity.datname = $1
		AND pid <> pg_backend_pid()`, name)
	if err != nil {
		return fmt.Errorf("failed to terminate connections: %v", err)
	}

	_, err = pool.Exec(ctx, fmt.Sprintf(`drop database %s`, name))
	if err != nil {
		return err
	}

	return nil
}

func runMigrations(_ context.Context, db *pgxpool.Pool) error {
	connStr := stdlib.RegisterConnConfig(db.Config().ConnConfig)
	defer stdlib.UnregisterConnConfig(connStr)

	sqlDB, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("failed to open SQL database: %v", err)
	}

	err = goose.SetDialect("pgx")
	if err != nil {
		return fmt.Errorf("failed to set dialect: %v", err)
	}
	goose.SetBaseFS(migrations.Embedded)

	err = goose.Up(sqlDB, ".")
	if err != nil {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	return nil
}
