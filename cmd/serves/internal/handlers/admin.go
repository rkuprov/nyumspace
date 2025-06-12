package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
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
			rest.InternalError(w, fmt.Errorf("failed to get users: %w", err))
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
			rest.InternalError(w, fmt.Errorf("failed to get homes: %w", err))
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
			rest.ValidationFailed(w, errors.New("userID is required"))
			return
		}

		resp, err := a.DeleteUser(r.Context(), &nyum.UserDeleteRequest{
			UserDeleteRequest: nyumpb.UserDeleteRequest{
				UserId: userID,
			},
		})
		if err != nil {
			rest.InternalError(w, fmt.Errorf("failed to delete user: %w", err))
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
			rest.ValidationFailed(w, errors.New("homeID is required"))
			return
		}

		resp, err := a.DeleteHome(r.Context(), &nyum.HomeDeleteRequest{
			HomeDeleteRequest: nyumpb.HomeDeleteRequest{
				HomeId: homeID,
			},
		})
		if err != nil {
			rest.InternalError(w, fmt.Errorf("failed to delete home: %w", err))
			return
		}

		rest.OK(w, resp)
	}
}
