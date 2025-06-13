package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rkuprov/nyumspace/cmd/serves/internal/homes"
	"github.com/rkuprov/nyumspace/pkg/auth"
	"github.com/rkuprov/nyumspace/pkg/gen/nyumpb"
	"github.com/rkuprov/nyumspace/pkg/nyum"
	"github.com/rkuprov/nyumspace/pkg/rest"
)

// Home Handlers

// CreateHome creates a handler for creating a new home
func CreateHome(h *homes.Homes) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := rest.ExtractPayload[nyumpb.HomeCreationRequest](r)
		if err != nil {
			rest.BadRequest(w, err)
			return
		}

		// Get the user ID from the header set by middleware
		userID := r.Header.Get(auth.UserIDHeader)
		if userID == "" {
			rest.Unauthorized(w, errors.New("user ID is required"))
			return
		}

		resp, err := h.CreateHome(r.Context(), &nyum.HomeCreationRequest{
			UserID: userID,
			HomeCreationRequest: nyumpb.HomeCreationRequest{
				Name:            req.Name,
				StreetAddress_1: req.StreetAddress_1,
				StreetAddress_2: req.StreetAddress_2,
				City:            req.City,
				State:           req.State,
				ZipCode:         req.ZipCode,
				Country:         req.Country,
				Description:     req.Description,
				Tags:            req.Tags,
				ImageUrl:        req.ImageUrl,
			},
		})
		if err != nil {
			rest.InternalError(w, fmt.Errorf("failed to create home: %w", err))
			return
		}

		rest.Created(w, resp)
	}
}

// GetHome creates a handler for retrieving a home by ID
func GetHome(h *homes.Homes) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		homeID := chi.URLParam(r, "homeID")
		if homeID == "" {
			rest.ValidationFailed(w, errors.New("homeID is required"))
			return
		}

		// Get the user ID from the header set by middleware
		userID := r.Header.Get(auth.UserIDHeader)

		resp, err := h.GetHome(r.Context(), &nyum.HomeRequest{
			HomeRequest: nyumpb.HomeRequest{
				HomeId:  homeID,
				OwnerId: userID,
			},
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Home not found: %v", err), http.StatusNotFound)
			return
		}

		rest.OK(w, resp)
	}
}

// UpdateHome creates a handler for updating a home
func UpdateHome(h *homes.Homes) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		homeID := chi.URLParam(r, "homeID")
		if homeID == "" {
			rest.ValidationFailed(w, errors.New("homeID is required"))
			return
		}

		req, err := rest.ExtractPayload[nyumpb.HomeUpdateRequest](r)
		if err != nil {
			rest.BadRequest(w, err)
			return
		}

		// Set the home ID from the URL parameter
		req.HomeId = homeID

		resp, err := h.UpdateHome(r.Context(), &nyum.HomeUpdateRequest{
			HomeUpdateRequest: nyumpb.HomeUpdateRequest{
				HomeId:          homeID,
				Name:            req.Name,
				OwnerId:         req.OwnerId,
				StreetAddress_1: req.StreetAddress_1,
				StreetAddress_2: req.StreetAddress_2,
				City:            req.City,
				State:           req.State,
				ZipCode:         req.ZipCode,
				Country:         req.Country,
				Description:     req.Description,
				Tags:            req.Tags,
				ImageUrl:        req.ImageUrl,
			},
		})
		if err != nil {
			rest.InternalError(w, fmt.Errorf("failed to update home: %w", err))
			return
		}

		rest.OK(w, resp)
	}
}

// DeleteHome creates a handler for deleting a home
func DeleteHome(h *homes.Homes) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		homeID := chi.URLParam(r, "homeID")
		if homeID == "" {
			rest.ValidationFailed(w, errors.New("homeID is required"))
			return
		}

		// Get the user ID from the header set by middleware
		userID := r.Header.Get(auth.UserIDHeader)

		resp, err := h.DeleteHome(r.Context(), &nyum.HomeDeleteRequest{
			UserID: userID,
			HomeDeleteRequest: nyumpb.HomeDeleteRequest{
				HomeId: homeID,
			},
		})
		if err != nil {
			if err.Error() == "unauthorized: user does not have permission to delete this home" {
				rest.Unauthorized(w, errors.New("you don't have permission to delete this home"))
				return
			}
			rest.InternalError(w, fmt.Errorf("failed to delete home: %w", err))
			return
		}

		rest.OK(w, resp)
	}
}

// GetAllHomesForUser creates a handler for retrieving all homes for a specific user
func GetAllHomesForUser(h *homes.Homes) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the user ID from the header set by middleware
		userID := r.Header.Get(auth.UserIDHeader)
		if userID == "" {
			rest.Unauthorized(w, errors.New("user ID is required"))
			return
		}

		resp, err := h.GetAllHomesForUser(r.Context(), userID)
		if err != nil {
			rest.InternalError(w, fmt.Errorf("failed to get homes for user: %w", err))
			return
		}

		rest.OK(w, resp)
	}
}
