package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/rkuprov/nyumspace/cmd/serves/internal/admin"
	"github.com/rkuprov/nyumspace/cmd/serves/internal/handlers"
	"github.com/rkuprov/nyumspace/cmd/serves/internal/homes"
	"github.com/rkuprov/nyumspace/cmd/serves/internal/users"
	"github.com/rkuprov/nyumspace/pkg/auth"
	"github.com/rkuprov/nyumspace/pkg/daemon"
	"github.com/rkuprov/nyumspace/pkg/store"
)

func main() {
	daemon.Run(func(ctx context.Context, d daemon.Daemon) error {
		// Initialize storage
		storageConfig := getStorageConfig()
		s, err := store.NewStore(storageConfig)
		if err != nil {
			log.Fatalf("Failed to initialize storage: %v", err)
		}

		setupRoutes(d, s)

		return d.Server.ListenAndServe()
	},
		daemon.WithAddress("localhost:3000"))
}

// getStorageConfig returns storage configuration based on environment variables
func getStorageConfig() *store.Config {
	provider := os.Getenv("STORAGE_PROVIDER")
	if provider == "" {
		provider = "localstack" // Default to localstack for development
	}

	switch provider {
	case "localstack":
		bucket := os.Getenv("S3_BUCKET")
		if bucket == "" {
			bucket = "nyumspace-images" // Default bucket name
		}
		return store.DefaultLocalStackConfig(bucket)
	case "s3":
		return &store.Config{
			Provider:    "s3",
			Region:      getEnvOrDefault("AWS_REGION", "us-east-1"),
			Bucket:      getEnvOrDefault("S3_BUCKET", "nyumspace-images"),
			AccessKeyID: os.Getenv("AWS_ACCESS_KEY_ID"),
			SecretKey:   os.Getenv("AWS_SECRET_ACCESS_KEY"),
		}
	default:
		log.Printf("Unknown storage provider '%s', defaulting to localstack", provider)
		return store.DefaultLocalStackConfig("nyumspace-images")
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func setupRoutes(d daemon.Daemon, s store.Store) {
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
	h := homes.NewHomes(&d, s) // Pass storage to homes service
	d.Router.Route("/api/portal/homes", func(r chi.Router) {
		//r.Use(m.Session)
		//r.Use(m.AllowUser)
		r.Get("/all", handlers.GetAllHomesForUser(h))  // Get all homes for the current user
		r.Post("/", handlers.CreateHome(h))            // Create a new home
		r.Get("/{home-id}", handlers.GetHome(h))       // Get home by ID
		r.Put("/{home-id}", handlers.UpdateHome(h))    // Update home
		r.Delete("/{home-id}", handlers.DeleteHome(h)) // Delete home

		// Image upload endpoints
		r.Post("/{home-id}/images/upload", handlers.UploadHomeImage(s))               // Direct image upload
		r.Post("/{home-id}/images/presigned", handlers.GeneratePresignedUploadURL(s)) // Generate presigned URL
		r.Delete("/{home-id}/images", handlers.DeleteHomeImage(s))                    // Delete image
	})

	// Set router to daemon
	d.Server.Handler = d.Router
}
