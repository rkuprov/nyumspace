package daemon

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/rkuprov/nyumspace/pkg/config"
)

type Daemon struct {
	DB      *pgxpool.Pool
	Server  *http.Server
	Router  *chi.Mux
	errChan chan error
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

	// Setup Daemon
	d, doneFuncs, err := newDaemon(ctx, cfg)
	if err != nil {
		errChan <- err

	}

	d.errChan = errChan

	d.Router = chi.NewRouter()
	d.Server.Handler = d.Router

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
	shutdownCtx, _ := context.WithTimeout(ctx, 30*time.Second)
	err = d.Server.Shutdown(shutdownCtx)
	if err != nil {
		log.Printf("Error shutting down HTTP server: %v\n", err)
	}

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

	// Initialize the HTTP server
	baseCtx, cancelInflightRequests := context.WithTimeout(ctx, 20*time.Second)
	httpServer := &http.Server{
		Addr: fmt.Sprintf("%s:%s", cfg.HTTPServer.Host, cfg.HTTPServer.Port),
		BaseContext: func(_ net.Listener) context.Context {
			return baseCtx
		},
	}
	log.Printf("Starting HTTP server on %s\n", httpServer.Addr)
	httpServer.RegisterOnShutdown(cancelInflightRequests)

	return Daemon{
		DB:     dbpool,
		Server: httpServer,
	}, doneFuncs, nil
}

func recoverDaemonPanic(errChan chan error) {
	if r := recover(); r != nil {
		log.Printf("Daemon panicked: %v", r)
		errChan <- fmt.Errorf("panic: %v", r)
	}
}

func (d *Daemon) RegisterError(err error) {
	if err == nil {
		return
	}
	d.errChan <- err
}
