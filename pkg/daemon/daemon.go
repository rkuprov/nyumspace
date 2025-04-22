package daemon

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type workFunc = func(context.Context) error

func Run(work workFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Channel to listen for errors
	errChan := make(chan error, 1)

	// Channel to listen for OS signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Start a goroutine to simulate some work
	go func() {
		err := work(ctx)
		if err != nil {
			errChan <- err
			return
		}
	}()

	// Main loop to handle quit conditions
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
