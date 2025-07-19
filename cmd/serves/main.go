package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/rkuprov/nyumspace/cmd/serves/internal/admin"
	"github.com/rkuprov/nyumspace/cmd/serves/internal/handlers"
	"github.com/rkuprov/nyumspace/cmd/serves/internal/homes"
	"github.com/rkuprov/nyumspace/cmd/serves/internal/users"
	"github.com/rkuprov/nyumspace/pkg/auth"
	"github.com/rkuprov/nyumspace/pkg/daemon"
)

func main() {
	daemon.Run(func(ctx context.Context, d daemon.Daemon) error {
		// Initialize storage

		setupRoutes(d)

		return d.Server.ListenAndServe()
	},
		daemon.WithAddress("localhost:3000"))
}

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
		r.Get("/users", handlers.GetAllUsers(&a))                   // Get all users
		r.Delete("/{userID}", handlers.AdminDeleteUser(&a))         // Delete user
		r.Get("/{userID}", handlers.AdminGetUser(&a))               // Get user by ID
		r.Get("/{userID}/homes", handlers.AdminGetHomesForUser(&a)) // Get homes by user ID

		r.Get("/homes", handlers.GetAllHomes(&a))                 // Get all homes
		r.Get("/homes/{homeID}", handlers.AdminGetHome(&a))       // Get home by ID
		r.Delete("/homes/{homeID}", handlers.AdminDeleteHome(&a)) // Delete home
	})

	// User routes
	u := users.NewUsers(&d)
	d.Router.Post("/register", handlers.RegisterUser(u)) // Register a new user
	d.Router.Post("/login", handlers.LoginUser(u))       // Login
	d.Router.Route("/api/portal", func(r chi.Router) {
		r.Use(m.Session)
		r.Use(m.AllowUser)
		r.Get("/api/portal", handlers.GetUser(u))       // Get user by ID
		r.Put("/api/portal", handlers.UpdateUser(u))    // Update user
		r.Delete("/api/portal", handlers.DeleteUser(u)) // Delete user
		r.Post("/api/logout", handlers.LogoutUser(u))   // Logout
	})

	// Home routes
	h := homes.NewHomes(&d) // Pass storage to homes service
	d.Router.Route("/api/portal/homes", func(r chi.Router) {
		//r.Use(m.Session)
		//r.Use(m.AllowUser)
		r.Get("/all", handlers.GetAllHomesForUser(h))  // Get all homes for the current user
		r.Post("/", handlers.CreateHome(h))            // Create a new home
		r.Get("/{home-id}", handlers.GetHome(h))       // Get home by ID
		r.Put("/{home-id}", handlers.UpdateHome(h))    // Update home
		r.Delete("/{home-id}", handlers.DeleteHome(h)) // Delete home

		// Image upload endpoints
		r.Post("/{home-id}/images/upload", handlers.UploadHomeImage(h))               // Direct image upload
		r.Post("/{home-id}/images/presigned", handlers.GeneratePresignedUploadURL(h)) // Generate presigned URL
		r.Delete("/{home-id}/images", handlers.DeleteHomeImage(h))                    // Delete image
	})

	// Set router to daemon
	d.Server.Handler = d.Router
}
