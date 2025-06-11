package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rkuprov/nyumspace/cmd/serves/internal/handlers"
	"github.com/rkuprov/nyumspace/cmd/serves/internal/homes"
	"github.com/rkuprov/nyumspace/cmd/serves/internal/users"
	"github.com/rkuprov/nyumspace/pkg/daemon"
)

func main() {
	daemon.Run(func(ctx context.Context, d daemon.Daemon) error {
		setupRoutes(d)

		return d.Server.ListenAndServe()
	},
		daemon.WithAddress("localhost:3000"))
}

func setupRoutes(d daemon.Daemon) {
	// Create a Chi router if not already created
	if d.Router == nil {
		d.Router = chi.NewRouter()
	}

	// Add middleware
	d.Router.Use(middleware.Logger)
	d.Router.Use(middleware.Recoverer)

	// Basic health check endpoint
	d.Router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Health check request received")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"message": "Service is running",
		})
	})

	// User routes
	u := users.NewUsers(&d)

	d.Router.Route("/api/users", func(r chi.Router) {
		r.Post("/register", handlers.RegisterUser(u)) // Register a new user
		r.Get("/{userID}", handlers.GetUser(u))       // Get user by ID
		r.Put("/{userID}", handlers.UpdateUser(u))    // Update user
		r.Delete("/{userID}", handlers.DeleteUser(u)) // Delete user
		r.Post("/login", handlers.LoginUser(u))       // Login
		r.Post("/logout", handlers.LogoutUser(u))     // Logout
		r.Get("/", handlers.GetAllUsers(u))           // Get all users
	})

	// Home routes
	h := homes.NewHomes(&d)

	d.Router.Route("/api/homes", func(r chi.Router) {
		r.Post("/", handlers.CreateHome(h))           // Create a new home
		r.Get("/{homeID}", handlers.GetHome(h))       // Get home by ID
		r.Put("/{homeID}", handlers.UpdateHome(h))    // Update home
		r.Delete("/{homeID}", handlers.DeleteHome(h)) // Delete home
	})

	// Set router to daemon
	d.Server.Handler = d.Router
}
