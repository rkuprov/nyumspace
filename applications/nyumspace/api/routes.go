package main

import (
	"github.com/go-chi/chi/v5"

	"github.com/rkuprov/nyumspace/applications/nyumspace/api/internal/handlers"
)

func SetupRoutes(router *chi.Mux) error {
	router.Get("/", handlers.Home())
	router.Route("/", func(r chi.Router) {
		r.Get("/health", handlers.HealthFunc())
		r.Get("/hello", handlers.Hello())
	})
	return nil
}
