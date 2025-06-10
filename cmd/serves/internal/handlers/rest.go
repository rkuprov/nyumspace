package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rkuprov/nyumspace/cmd/serves/internal/sql"
	"github.com/rkuprov/nyumspace/pkg/gen/nyumpb"
	"golang.org/x/crypto/bcrypt"
)

// Response is a generic response structure for REST endpoints
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// SendJSON sends a JSON response with the provided status code
func SendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// SendError sends an error response with the provided status code
func SendError(w http.ResponseWriter, status int, message string) {
	SendJSON(w, status, Response{
		Success: false,
		Error:   message,
	})
}

// User Handlers

// RegisterUser creates a handler for user registration
func RegisterUser(db *pgxpool.Pool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var req nyumpb.UserRegistrationRequest
		out, _ := io.ReadAll(r.Body)
		err := json.Unmarshal(out, &req)
		if err != nil {
			SendError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Validate required fields
		if req.Username == "" || req.Email == "" || req.Password == "" {
			SendError(w, http.StatusBadRequest, "Username, email, and password are required")
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			SendError(w, http.StatusInternalServerError, "Failed to hash password")
			return
		}

		id := uuid.NewString()
		row := db.QueryRow(r.Context(), sql.RegisterUser, id, req.Username, req.Email, string(hash))
		if err = row.Scan(&id); err != nil {
			SendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to register user: %v", err))
			return
		}

		response := nyumpb.UserRegistrationResponse{
			Success: true,
			Message: fmt.Sprintf("User %s registered successfully with ID: %s", req.Username, id),
			UserId:  id,
		}

		SendJSON(w, http.StatusCreated, response)
	}
}

// GetUser creates a handler for retrieving a user by ID
func GetUser(db *pgxpool.Pool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userID")
		if userID == "" {
			SendError(w, http.StatusBadRequest, "User ID is required")
			return
		}
		row := db.QueryRow(r.Context(), sql.GetUser, userID)
		var id, name, email string
		if err := row.Scan(&id, &name, &email); err != nil {
			SendError(w, http.StatusNotFound, fmt.Sprintf("User not found: %v", err))
			return
		}

		response := nyumpb.UserResponse{
			UserId:   id,
			Username: name,
			Email:    email,
		}

		SendJSON(w, http.StatusOK, response)
	}
}

// UpdateUser creates a handler for updating a user
func UpdateUser(db *pgxpool.Pool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userID")
		if userID == "" {
			SendError(w, http.StatusBadRequest, "User ID is required")
			return
		}

		var req nyumpb.UserUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Set the user ID from the URL parameter
		req.UserId = userID

		row := db.QueryRow(r.Context(), sql.UpdateUser, req.UserId, req.Username, req.Email)
		var id int32
		if err := row.Scan(&id); err != nil {
			SendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update user: %v", err))
			return
		}

		response := nyumpb.UserUpdateResponse{
			Success: true,
			Message: fmt.Sprintf("User with ID %s updated successfully", userID),
		}

		SendJSON(w, http.StatusOK, response)
	}
}

// DeleteUser creates a handler for deleting a user
func DeleteUser(db *pgxpool.Pool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userID")
		if userID == "" {
			SendError(w, http.StatusBadRequest, "User ID is required")
			return
		}

		_, err := db.Exec(r.Context(), sql.DeleteUser, userID)
		if err != nil {
			SendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete user: %v", err))
			return
		}

		response := nyumpb.UserDeleteResponse{
			Success: true,
			Message: fmt.Sprintf("User with ID %s deleted successfully", userID),
		}

		SendJSON(w, http.StatusOK, response)
	}
}

// LoginUser creates a handler for user login
func LoginUser(db *pgxpool.Pool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var req nyumpb.UserLoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Validate required fields
		if req.Email == "" || req.Password == "" {
			SendError(w, http.StatusBadRequest, "Email and password are required")
			return
		}

		var userID, hashedPassword string
		row := db.QueryRow(r.Context(), sql.GetUserByEmail, req.Email)
		if err := row.Scan(&userID, &hashedPassword); err != nil {
			SendError(w, http.StatusUnauthorized, "Invalid email or password")
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
			SendError(w, http.StatusUnauthorized, "Invalid email or password")
			return
		}

		// Generate session token
		sessionToken := uuid.New().String()

		expiresAt := time.Now().Add(24 * time.Hour)
		_, err := db.Exec(r.Context(), sql.CreateSession, sessionToken, userID, expiresAt)
		if err != nil {
			SendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create session: %v", err))
			return
		}

		response := nyumpb.UserLoginResponse{
			Success:      true,
			Message:      "Login successful",
			SessionToken: sessionToken,
			UserId:       userID,
		}

		SendJSON(w, http.StatusOK, response)
	}
}

// LogoutUser creates a handler for user logout
func LogoutUser(db *pgxpool.Pool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get session token from Authorization header
		sessionToken := r.Header.Get("Authorization")
		if sessionToken == "" {
			SendError(w, http.StatusBadRequest, "Session token is required")
			return
		}

		// Remove "Bearer " prefix if present
		if len(sessionToken) > 7 && sessionToken[:7] == "Bearer " {
			sessionToken = sessionToken[7:]
		}

		_, err := db.Exec(r.Context(), sql.DeleteSession, sessionToken)
		if err != nil {
			SendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete session: %v", err))
			return
		}

		response := nyumpb.UserLogoutResponse{
			Success: true,
			Message: "Logout successful",
		}

		SendJSON(w, http.StatusOK, response)
	}
}

// Home Handlers

// CreateHome creates a handler for creating a new home
func CreateHome(db *pgxpool.Pool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var req nyumpb.HomeCreationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Validate required fields
		if req.OwnerId == "" || req.Name == "" {
			SendError(w, http.StatusBadRequest, "Owner ID and name are required")
			return
		}

		homeID := uuid.New().String()

		_, err := db.Exec(r.Context(), sql.AddHomeSQL,
			homeID,
			req.OwnerId,
			req.Name,
			req.StreetAddress_1,
			req.StreetAddress_2,
			req.City,
			req.State,
			req.ZipCode,
			req.Country,
			req.Description,
			req.Tags,
			req.ImageUrl,
		)
		if err != nil {
			SendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create home: %v", err))
			return
		}

		response := nyumpb.HomeCreationResponse{
			HomeId:  homeID,
			Success: true,
			Message: fmt.Sprintf("Home '%s' created successfully", req.Name),
		}

		SendJSON(w, http.StatusCreated, response)
	}
}

// GetHome creates a handler for retrieving a home by ID
func GetHome(db *pgxpool.Pool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		homeID := chi.URLParam(r, "homeID")
		if homeID == "" {
			SendError(w, http.StatusBadRequest, "Home ID is required")
			return
		}

		row := db.QueryRow(r.Context(), sql.GetHomeSQL, homeID)

		var (
			id             string
			ownerID        string
			name           string
			streetAddress1 string
			streetAddress2 string
			city           string
			state          string
			zipCode        string
			country        string
			description    string
			tags           []string
			imageURL       string
			createdAt      time.Time
			updatedAt      time.Time
		)

		if err := row.Scan(&id, &ownerID, &name, &streetAddress1, &streetAddress2, &city, &state,
			&zipCode, &country, &description, &tags, &imageURL, &createdAt, &updatedAt); err != nil {
			SendError(w, http.StatusNotFound, fmt.Sprintf("Home not found: %v", err))
			return
		}

		response := nyumpb.HomeResponse{
			HomeId:          id,
			OwnerId:         ownerID,
			Name:            name,
			Description:     description,
			StreetAddress_1: streetAddress1,
			StreetAddress_2: streetAddress2,
			City:            city,
			State:           state,
			ZipCode:         zipCode,
			Country:         country,
			ImageUrl:        imageURL,
			Tags:            tags,
			CreatedAt:       createdAt.Format(time.RFC3339),
			UpdatedAt:       updatedAt.Format(time.RFC3339),
		}

		SendJSON(w, http.StatusOK, response)
	}
}

// UpdateHome creates a handler for updating a home
func UpdateHome(db *pgxpool.Pool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		homeID := chi.URLParam(r, "homeID")
		if homeID == "" {
			SendError(w, http.StatusBadRequest, "Home ID is required")
			return
		}

		var req nyumpb.HomeUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Not implemented in the original handler
		SendError(w, http.StatusNotImplemented, "Update home functionality not implemented yet")
	}
}

// DeleteHome creates a handler for deleting a home
func DeleteHome(db *pgxpool.Pool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		homeID := chi.URLParam(r, "homeID")
		if homeID == "" {
			SendError(w, http.StatusBadRequest, "Home ID is required")
			return
		}

		row := db.QueryRow(r.Context(), sql.DeleteHomeSQL, homeID)

		var id string
		if err := row.Scan(&id); err != nil {
			SendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete home: %v", err))
			return
		}

		response := nyumpb.HomeDeleteResponse{
			Success: true,
			Message: fmt.Sprintf("Home with ID %s deleted successfully", homeID),
		}

		SendJSON(w, http.StatusOK, response)
	}
}
