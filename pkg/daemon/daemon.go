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
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/rkuprov/nyumspace/pkg/config"
	"github.com/rkuprov/nyumspace/pkg/storage"
)

type Daemon struct {
	DB      *pgxpool.Pool
	Server  *http.Server
	Router  *chi.Mux
	Storage *storage.Client // Assuming storage.Client is defined elsewhere
	errChan chan error
}

type workFunc = func(context.Context, Daemon) error

func Run(work workFunc, opts ...DaemonOpt) {
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
	d, doneFuncs, err := newDaemon(ctx, cfg, opts...)
	if err != nil {
		errChan <- err

	}

	d.errChan = errChan

	go func() {
		defer recoverDaemonPanic(errChan)
		closeErr := work(ctx, d)
		if closeErr != nil {
			errChan <- closeErr
			return
		}
	}()

	select {
	case sig := <-signalChan:
		log.Printf("Received signal: %s, shutting down...\n", sig)
	case err = <-errChan:
		log.Printf("Error occurred: %v, shutting down...\n", err)
	case <-ctx.Done():
		log.Println("Context canceled, shutting down...")
	}

	log.Println("Gracefully shutting down...")
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	err = d.Server.Shutdown(shutdownCtx)
	if err != nil {
		log.Printf("Error shutting down HTTP server: %v\n", err)
	}

	for _, done := range doneFuncs {
		done()
	}

	time.Sleep(time.Second * 5)
	cancel()
	log.Println("Daemon stopped.")
}

type closers = []func()

func newDaemon(ctx context.Context, cfg config.Cfg, opts ...DaemonOpt) (Daemon, closers, error) {
	var doneFuncs closers
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.PG.User, cfg.PG.Password, cfg.PG.Host, cfg.PG.Port, cfg.PG.DbName)

	dbpool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return Daemon{}, nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Verify the connection
	if err = dbpool.Ping(ctx); err != nil {
		dbpool.Close()
		return Daemon{}, nil, fmt.Errorf("unable to ping database: %w", err)
	}

	store, err := storage.NewStorageClient(ctx, cfg.S3Aws)
	if err != nil {
		return Daemon{}, nil, fmt.Errorf("unable to create storage client: %w", err)
	}

	doneFuncs = append(doneFuncs, dbpool.Close)
	log.Println("Successfully connected to PostgreSQL")

	// Initialize the HTTP server
	baseCtx, cancelInflightRequests := context.WithCancel(ctx)
	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Use(middleware.Logger)
	httpServer := &http.Server{
		Addr: fmt.Sprintf("%s:%s", cfg.HTTPServer.Host, cfg.HTTPServer.Port),
		BaseContext: func(_ net.Listener) context.Context {
			return baseCtx
		},
		Handler:     router,
		ReadTimeout: 4 * time.Second,
	}
	httpServer.RegisterOnShutdown(cancelInflightRequests)

	d := Daemon{
		DB:      dbpool,
		Server:  httpServer,
		Router:  router,
		Storage: store,
	}

	for _, opt := range opts {
		if err := opt(&d); err != nil {
			return Daemon{}, nil, fmt.Errorf("error applying daemon option: %w", err)
		}
	}

	log.Printf("Starting HTTP server on %s\n", httpServer.Addr)
	return d, doneFuncs, nil
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
