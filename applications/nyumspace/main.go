package main

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/rkuprov/nyumspace/applications/nyumspace/routes"
)

func main() {
	r := chi.NewRouter()
	ctx := context.Background()

	r.Get("/hello", routes.Hello(ctx))

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}
}
