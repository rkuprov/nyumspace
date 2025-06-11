package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rkuprov/nyumspace/cmd/serves/internal/sql"
	"github.com/rkuprov/nyumspace/cmd/serves/internal/users"
	"github.com/rkuprov/nyumspace/pkg/gen/nyumpb"
	"github.com/rkuprov/nyumspace/pkg/nyum"
)

// User Handlers

// RegisterUser creates a handler for user registration
func RegisterUser(u *users.Users) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := extractPayload[nyumpb.UserRegistrationRequest](r)
		if err != nil {
			BadRequest(w, err)
			return
		}

		// Validate required fields
		if req.Username == "" || req.Email == "" || req.Password == "" {
			ValidationFailed(w, errors.New("must have username, email, and password"))
			return
		}

		resp, err := u.RegisterUser(r.Context(), &nyum.UserRegistrationRequest{
			UserRegistrationRequest: nyumpb.UserRegistrationRequest{
				Username: req.Username,
				Password: req.Password,
				Email:    req.Email,
			},
		})
		if err != nil {
			InternalError(w, fmt.Errorf("failed to register user: %w", err))
			return
		}

		Created(w, resp)
	}
}

// GetUser creates a handler for retrieving a user by ID
func GetUser(u *users.Users) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userID")
		if userID == "" {
			ValidationFailed(w, errors.New("userID is required"))
			return
		}

		resp, err := u.GetUser(r.Context(), &nyum.UserRequest{
			UserRequest: nyumpb.UserRequest{
				UserId: userID,
			},
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("User not found: %v", err), http.StatusNotFound)
			return
		}

		OK(w, resp)
	}
}

// UpdateUser creates a handler for updating a user
func UpdateUser(u *users.Users) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userID")
		if userID == "" {
			ValidationFailed(w, errors.New("userID is required"))
			return
		}

		req, err := extractPayload[nyumpb.UserUpdateRequest](r)
		if err != nil {
			BadRequest(w, err)
			return
		}

		// Set the user ID from the URL parameter
		req.UserId = userID

		resp, err := u.UpdateUser(r.Context(), &nyum.UserUpdateRequest{
			UserUpdateRequest: nyumpb.UserUpdateRequest{
				UserId:   userID,
				Username: req.Username,
				Email:    req.Email,
			},
		})
		if err != nil {
			InternalError(w, fmt.Errorf("failed to update user: %w", err))
			return
		}

		OK(w, resp)
	}
}

// DeleteUser creates a handler for deleting a user
func DeleteUser(u *users.Users) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userID")
		if userID == "" {
			ValidationFailed(w, errors.New("userID is required"))
			return
		}

		resp, err := u.DeleteUser(r.Context(), &nyum.UserDeleteRequest{
			UserDeleteRequest: nyumpb.UserDeleteRequest{
				UserId: userID,
			},
		})
		if err != nil {
			InternalError(w, fmt.Errorf("failed to delete user: %w", err))
			return
		}

		OK(w, resp)
	}
}

// LoginUser creates a handler for user login
func LoginUser(u *users.Users) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := extractPayload[nyumpb.UserLoginRequest](r)
		if err != nil {
			BadRequest(w, err)
			return
		}

		// Validate required fields
		if req.Email == "" || req.Password == "" {
			ValidationFailed(w, errors.New("email and password are required"))
			return
		}

		resp, err := u.LoginUser(r.Context(), &nyum.UserLoginRequest{
			UserLoginRequest: nyumpb.UserLoginRequest{
				Email:    req.Email,
				Password: req.Password,
			},
		})
		if err != nil {
			if err.Error() == "invalid email or password" {
				http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			} else {
				InternalError(w, fmt.Errorf("login failed: %w", err))
			}
			return
		}

		OK(w, resp)
	}
}

// LogoutUser creates a handler for user logout
func LogoutUser(u *users.Users) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get session token from Authorization header
		sessionToken := r.Header.Get("Authorization")
		if sessionToken == "" {
			Unauthorized(w, errors.New("missing Authorization header"))
			return
		}

		// Remove "Bearer " prefix if present
		if len(sessionToken) > 7 && sessionToken[:7] == "Bearer " {
			sessionToken = sessionToken[7:]
		}

		resp, err := u.LogoutUser(r.Context(), &nyum.UserLogoutRequest{
			UserLogoutRequest: nyumpb.UserLogoutRequest{
				SessionToken: sessionToken,
			},
		})
		if err != nil {
			InternalError(w, fmt.Errorf("failed to logout: %w", err))
			return
		}

		OK(w, resp)
	}
}

// Home Handlers

// CreateHome creates a handler for creating a new home
func CreateHome(db *pgxpool.Pool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := extractPayload[nyumpb.HomeCreationRequest](r)
		if err != nil {
			BadRequest(w, err)
			return
		}

		// Validate required fields
		if req.OwnerId == "" || req.Name == "" {
			ValidationFailed(w, errors.New("owner and name are required"))
			return
		}

		homeID := uuid.New().String()

		_, err = db.Exec(r.Context(), sql.AddHomeSQL,
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
			InternalError(w, fmt.Errorf("failed to create home: %w", err))
			return
		}

		response := nyumpb.HomeCreationResponse{
			HomeId:  homeID,
			Message: fmt.Sprintf("Home '%s' created successfully", req.Name),
		}

		Created(w, response)
	}
}

// GetHome creates a handler for retrieving a home by ID
func GetHome(db *pgxpool.Pool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		homeID := chi.URLParam(r, "homeID")
		if homeID == "" {
			ValidationFailed(w, errors.New("homeID is required"))
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
			http.Error(w, fmt.Sprintf("Home not found: %v", err), http.StatusNotFound)
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

		OK(w, response)
	}
}

// UpdateHome creates a handler for updating a home
func UpdateHome(db *pgxpool.Pool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		homeID := chi.URLParam(r, "homeID")
		if homeID == "" {
			ValidationFailed(w, errors.New("homeID is required"))
			return
		}

		_, err := extractPayload[nyumpb.HomeUpdateRequest](r)
		if err != nil {
			BadRequest(w, err)
			return
		}

		// Not implemented in the original handler
		http.Error(w, "Update home functionality not implemented yet", http.StatusNotImplemented)
	}
}

// DeleteHome creates a handler for deleting a home
func DeleteHome(db *pgxpool.Pool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		homeID := chi.URLParam(r, "homeID")
		if homeID == "" {
			ValidationFailed(w, errors.New("homeID is required"))
			return
		}

		row := db.QueryRow(r.Context(), sql.DeleteHomeSQL, homeID)

		var id string
		if err := row.Scan(&id); err != nil {
			InternalError(w, fmt.Errorf("failed to delete home: %w", err))
			return
		}

		response := nyumpb.HomeDeleteResponse{
			Message: fmt.Sprintf("Home with ID %s deleted successfully", homeID),
		}

		OK(w, response)
	}
}
