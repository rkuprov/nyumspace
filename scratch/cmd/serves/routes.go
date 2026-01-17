package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/rkuprov/nyumspace/scratch/cmd/serves/internal/admin"
	handlers2 "github.com/rkuprov/nyumspace/scratch/cmd/serves/internal/handlers"
	"github.com/rkuprov/nyumspace/scratch/cmd/serves/internal/homes"
	"github.com/rkuprov/nyumspace/scratch/cmd/serves/internal/users"
	"github.com/rkuprov/nyumspace/scratch/pkg/app/auth"
	"github.com/rkuprov/nyumspace/scratch/pkg/app/daemon"
)

func setupRoutes(d daemon.Daemon) {
	// Create a Chi router if not already created
	if d.Router == nil {
		panic("router not initialized")
	}

	m := auth.NewMiddleware(&d)

	// Basic health check endpoint
	d.Router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Health check request received")
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"message": "Service is running",
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Admin routes
	a := admin.NewAdmin(d)
	d.Router.Route("/admin", func(r chi.Router) {
		r.Use(m.Session)
		r.Use(m.AllowUser)
		r.Get("/users", handlers2.GetAllUsers(&a))                   // Get all users
		r.Delete("/{userID}", handlers2.AdminDeleteUser(&a))         // Delete user
		r.Get("/{userID}", handlers2.AdminGetUser(&a))               // Get user by ID
		r.Get("/{userID}/homes", handlers2.AdminGetHomesForUser(&a)) // Get homes by user ID

		r.Get("/homes", handlers2.GetAllHomes(&a))                 // Get all homes
		r.Get("/homes/{homeID}", handlers2.AdminGetHome(&a))       // Get home by ID
		r.Delete("/homes/{homeID}", handlers2.AdminDeleteHome(&a)) // Delete home
	})

	// User routes
	u := users.NewUsers(&d)
	d.Router.Post("/register", handlers2.RegisterUser(u)) // Register a new user
	d.Router.Post("/login", handlers2.LoginUser(u))       // Login
	d.Router.Route("/api/portal", func(r chi.Router) {
		r.Use(m.Session)
		r.Use(m.AllowUser)
		r.Get("/api/portal", handlers2.GetUser(u))       // Get user by ID
		r.Put("/api/portal", handlers2.UpdateUser(u))    // Update user
		r.Delete("/api/portal", handlers2.DeleteUser(u)) // Delete user
		r.Post("/api/logout", handlers2.LogoutUser(u))   // Logout
	})

	// Home routes
	h := homes.NewHomes(&d) // Pass storage to homes service
	d.Router.Route("/api/portal/homes", func(r chi.Router) {
		r.Use(m.Session)
		r.Use(m.AllowUser)
		r.Get("/all", handlers2.GetAllHomesForUser(h))  // Get all homes for the current user
		r.Post("/", handlers2.CreateHome(h))            // Create a new home
		r.Get("/{home-id}", handlers2.GetHome(h))       // Get home by ID
		r.Put("/{home-id}", handlers2.UpdateHome(h))    // Update home
		r.Delete("/{home-id}", handlers2.DeleteHome(h)) // Delete home

		// Image upload endpoints
		r.Post("/{home-id}/images", handlers2.UploadHomeImage(h))             // Direct image upload
		r.Get("/{home-id}/images/{imageID}", handlers2.GetHomeImage(h))       // Get home image by ID
		r.Delete("/{home-id}/images/{imageID}", handlers2.DeleteHomeImage(h)) // Delete home image by ID
	})

	// Set router to daemon
	d.Server.Handler = d.Router
}
