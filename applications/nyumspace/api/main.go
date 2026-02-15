package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/rkuprov/nyumspace/applications/nyumspace/api/internal/handlers"
)

func main() {
	r := chi.NewRouter()
	err := SetupRoutes(r)
	if err != nil {
		panic(err)
	}

	err = http.ListenAndServe(":8000", r)
	if err != nil {
		panic(err)
	}
}
func SetupRoutes(router *chi.Mux) error {
	router.Get("/", handlers.Home())
	router.Route("/", func(r chi.Router) {
		r.Get("/health", handlers.HealthFunc())
		r.Get("/hello", handlers.Hello())
	})
	return nil
}
