package daemon

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
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

	d, doneFuncs, err := newDaemon(ctx, cfg)
	if err != nil {
		errChan <- err

	}

	// Start a goroutine to simulate some work
	go func() {
		defer recoverDaemonPanic(errChan)
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

	log.Println("Gracefully shutting down...")
	for _, done := range doneFuncs {
		done()
	}

	log.Println("Daemon stopped.")
}

type closers = []func()

func newDaemon(ctx context.Context, cfg config.Cfg) (Daemon, closers, error) {
	var doneFuncs closers
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.PG.User, cfg.PG.Password, cfg.PG.Host, cfg.PG.Port, cfg.PG.DbName)

	dbpool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return Daemon{}, nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Verify the connection
	if err := dbpool.Ping(ctx); err != nil {
		dbpool.Close()
		return Daemon{}, nil, fmt.Errorf("unable to ping database: %w", err)
	}

	doneFuncs = append(doneFuncs, dbpool.Close)
	log.Println("Successfully connected to PostgreSQL")

	return Daemon{DB: dbpool}, doneFuncs, nil
}

func recoverDaemonPanic(errChan chan error) {
	if r := recover(); r != nil {
		log.Printf("Daemon panicked: %v", r)
		errChan <- fmt.Errorf("panic: %v", r)
	}
}
