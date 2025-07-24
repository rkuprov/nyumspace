package main

import (
	"context"

	"github.com/rkuprov/nyumspace/pkg/daemon"
)

func main() {
	daemon.Run(func(ctx context.Context, d daemon.Daemon) error {
		setupRoutes(d)
		return d.Server.ListenAndServe()
	},
		daemon.WithAddress("localhost:3000"))
}
