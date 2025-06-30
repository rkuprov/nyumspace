package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/rkuprov/nyumspace/cmd/serves/internal/admin"
	"github.com/rkuprov/nyumspace/pkg/gen/nyumpb"
	"github.com/rkuprov/nyumspace/pkg/nyum"
	"github.com/rkuprov/nyumspace/pkg/rest"
)

// GetAllUsers creates a handler for retrieving all users
func GetAllUsers(a *admin.Admin) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := a.GetAllUsers(r.Context())
		if err != nil {
			rest.ErrInternal(w, fmt.Errorf("failed to get users: %w", err))
			return
		}

		rest.OK(w, resp...)
	}
}

// GetAllHomes creates a handler for retrieving all homes
func GetAllHomes(a *admin.Admin) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := a.GetAllHomes(r.Context())
		if err != nil {
			rest.ErrInternal(w, fmt.Errorf("failed to get homes: %w", err))
			return
		}

		rest.OK(w, resp...)
	}
}

// AdminDeleteUser creates a handler for deleting a user
func AdminDeleteUser(a *admin.Admin) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userID")
		if userID == "" {
			rest.ErrValidation(w, errors.New("userID is required"))
			return
		}

		resp, err := a.DeleteUser(r.Context(), &nyum.UserDeleteRequest{
			UserDeleteRequest: nyumpb.UserDeleteRequest{
				UserId: userID,
			},
		})
		if err != nil {
			rest.ErrInternal(w, fmt.Errorf("failed to delete user: %w", err))
			return
		}

		rest.OK(w, resp)
	}
}

// AdminDeleteHome creates a handler for deleting a home
func AdminDeleteHome(a *admin.Admin) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		homeID := chi.URLParam(r, "homeID")
		if homeID == "" {
			rest.ErrValidation(w, errors.New("homeID is required"))
			return
		}

		resp, err := a.DeleteHome(r.Context(), &nyum.HomeDeleteRequest{
			HomeDeleteRequest: nyumpb.HomeDeleteRequest{
				HomeId: homeID,
			},
		})
		if err != nil {
			rest.ErrInternal(w, fmt.Errorf("failed to delete home: %w", err))
			return
		}

		rest.OK(w, resp)
	}
}

// AdminGetHome creates a handler for retrieving a home by ID
func AdminGetHome(a *admin.Admin) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		homeID := chi.URLParam(r, "home-id")
		if homeID == "" {
			rest.ErrValidation(w, errors.New("homeID is required"))
			return
		}

		resp, err := a.GetHome(r.Context(), &nyum.HomeRequest{
			HomeRequest: nyumpb.HomeRequest{
				HomeId: homeID,
			},
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				rest.ErrNotFound(w, fmt.Errorf("home not found"))
				return
			}
			rest.ErrInternal(w, fmt.Errorf("failed to get home: %w", err))
			return
		}

		rest.OK(w, resp)
	}
}

// AdminGetHomesForUser creates a handler for retrieving all homes for a specific user
func AdminGetHomesForUser(a *admin.Admin) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "user-id")
		if userID == "" {
			rest.ErrValidation(w, errors.New("userID is required"))
			return
		}

		resp, err := a.GetHomesForUser(r.Context(), &nyum.UserHomesRequest{
			UserId: userID,
		})
		if err != nil {
			rest.ErrInternal(w, fmt.Errorf("failed to get homes for user: %w", err))
			return
		}

		rest.OK(w, resp...)
	}
}

// AdminGetUser creates a handler for retrieving a user by ID
func AdminGetUser(a *admin.Admin) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userID")
		if userID == "" {
			rest.ErrValidation(w, errors.New("userID is required"))
			return
		}

		resp, err := a.GetUser(r.Context(), &nyum.UserRequest{
			UserRequest: nyumpb.UserRequest{
				UserId: userID,
			},
		})
		if err != nil {
			rest.ErrInternal(w, fmt.Errorf("failed to get user: %w", err))
			return
		}
		if resp == nil {
			rest.NotFound(w, fmt.Sprintf("user %s not found", userID))
			return
		}

		rest.OK(w, resp)
	}
}
