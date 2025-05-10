package main

import (
	"context"
	"log"
	"net/http"

	"github.com/rkuprov/nyumspace/cmd/servers/internal/handlers"
	"github.com/rkuprov/nyumspace/pkg/daemon"
	"github.com/rkuprov/nyumspace/pkg/gen/nyumpb/nyumpbconnect"
)

func main() {
	daemon.Run(func(ctx context.Context, d daemon.Daemon) error {
		svc := handlers.NewServerHandler()
		d.Router.Handle(nyumpbconnect.NewServerServiceHandler(svc))
		d.Router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			log.Println("Request Received")
			w.Write([]byte("Hello, world!"))
		})

		return d.Server.ListenAndServe()
	})
}
