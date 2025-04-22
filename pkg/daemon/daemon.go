package daemon

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"log"
	"nyum/migrations"
	"nyum/pkg/config"
	"os"
	"os/signal"
	"syscall"
)

type Daemon struct {
	DB *pgxpool.Pool
}

type workFunc = func(context.Context, Daemon) error

func Run(work workFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Channel to listen for errors
	errChan := make(chan error, 1)

	// Channel to listen for OS signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	cfg, err := config.NewConfig()
	if err != nil {
		errChan <- err
	}

	err = applyDBMigrations(ctx, cfg)
	if err != nil {
		errChan <- err
	}

	d, err := newDaemon(ctx, cfg)
	if err != nil {
		errChan <- err

	}

	// Start a goroutine to simulate some work
	go func() {
		err := work(ctx, d)
		if err != nil {
			errChan <- err
			return
		}
	}()

	select {
	case sig := <-signalChan:
		log.Printf("Received signal: %s, shutting down...\n", sig)
	case err := <-errChan:
		log.Printf("Error occurred: %v, shutting down...\n", err)
	case <-ctx.Done():
		log.Println("Context canceled, shutting down...")
	}

	log.Println("Daemon stopped.")
}

func newDaemon(ctx context.Context, cfg config.Cfg) (Daemon, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.PG.User, cfg.PG.Password, cfg.PG.Host, cfg.PG.Port, cfg.PG.DbName)

	dbpool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return Daemon{}, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Verify the connection
	if err := dbpool.Ping(ctx); err != nil {
		dbpool.Close()
		return Daemon{}, fmt.Errorf("unable to ping database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL")

	return Daemon{DB: dbpool}, nil
}

func applyDBMigrations(ctx context.Context, cfg config.Cfg) error {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.PG.Host, cfg.PG.Port, cfg.PG.User, cfg.PG.Password, cfg.PG.DbName)

	fmt.Println("opening db connection for migrations")
	// must use "pgx" driver for goose
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		fmt.Println("unable to open db connection")
		return err
	}
	defer db.Close()

	goose.SetBaseFS(migrations.EmbeddedMigrations)

	err = goose.SetDialect("postgres")
	if err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	// Run goose migrations
	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	log.Println("Migrations applied successfully")
	return nil
}
