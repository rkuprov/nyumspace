package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/rkuprov/nyumspace/scratch/cmd/serves/internal/homes"
	"github.com/rkuprov/nyumspace/scratch/pkg/api/rest"
	"github.com/rkuprov/nyumspace/scratch/pkg/app/auth"
	"github.com/rkuprov/nyumspace/scratch/pkg/gen/nyumpb"
	"github.com/rkuprov/nyumspace/scratch/pkg/nyum"
)

// Home Handlers

// CreateHome creates a handler for creating a new home
func CreateHome(h *homes.Homes) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := rest.ExtractPayload[nyumpb.HomeCreationRequest](r)
		if err != nil {
			rest.ErrBadRequest(w, err)
			return
		}

		// Get the user ID from the header set by middleware
		userID := r.Header.Get(auth.UserIDHeader)
		if userID == "" {
			rest.ErrUnauthorized(w, errors.New("user ID is required"))
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
			rest.ErrInternal(w, fmt.Errorf("failed to create home: %w", err))
			return
		}

		rest.Created(w, resp)
	}
}

// GetHome creates a handler for retrieving a home by ID
func GetHome(h *homes.Homes) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		homeID := chi.URLParam(r, "home-id")
		if homeID == "" {
			rest.ErrValidation(w, errors.New("homeID is required"))
			return
		}

		// Get the user ID from the header set by middleware
		userID := r.Header.Get(auth.UserIDHeader)

		resp, err := h.GetHome(r.Context(), &nyum.HomeRequest{
			HomeRequest: nyumpb.HomeRequest{
				HomeId: homeID,
			},
		})
		if err != nil {
			rest.ErrInternal(w, err)
			return
		}

		if resp == nil {
			rest.NotFound(w, fmt.Sprintf("home with ID %s not found", homeID))
			return
		}

		// Authorization check: Ensure the requesting user is the owner of the home
		if resp.GetOwnerId() != "" && resp.GetOwnerId() != userID {
			rest.ErrUnauthorized(w, errors.New("user unauthorized to access this home"))
			return
		}

		rest.OK(w, resp)
	}
}

// UpdateHome creates a handler for updating a home
func UpdateHome(h *homes.Homes) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		homeID := chi.URLParam(r, "home-id")
		if homeID == "" {
			rest.ErrValidation(w, errors.New("home-id is required"))
			return
		}

		req, err := rest.ExtractPayload[nyumpb.HomeUpdateRequest](r)
		if err != nil {
			rest.ErrBadRequest(w, err)
			return
		}

		// Get the user ID from the header set by middleware
		userID := r.Header.Get(auth.UserIDHeader)
		if userID == "" {
			rest.ErrUnauthorized(w, errors.New("user ID is required"))
			return
		}

		got, err := h.GetHome(r.Context(), &nyum.HomeRequest{
			HomeRequest: nyumpb.HomeRequest{
				HomeId: homeID,
			},
		})
		if err != nil {
			rest.ErrInternal(w, fmt.Errorf("failed to get home: %w", err))
			return
		}
		if got == nil {
			rest.ErrNotFound(w, fmt.Errorf("home with ID %s not found", homeID))
			return
		}
		if got.GetOwnerId() != userID {
			rest.ErrUnauthorized(w, errors.New("user unauthorized to update this home"))
			return
		}

		resp, err := h.UpdateHome(r.Context(), &nyum.HomeUpdateRequest{
			UserID: userID,
			HomeID: homeID,
			HomeUpdateRequest: nyumpb.HomeUpdateRequest{
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
			if errors.Is(err, homes.ErrNotFound) {
				rest.ErrNotFound(w, fmt.Errorf("home with ID %s not found", homeID))
				return
			}
			rest.ErrInternal(w, fmt.Errorf("failed to update home: %w", err))
			return
		}

		rest.OK(w, resp)
	}
}

// DeleteHome creates a handler for deleting a home
func DeleteHome(h *homes.Homes) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		homeID := chi.URLParam(r, "home-id")
		if homeID == "" {
			rest.ErrValidation(w, errors.New("home ID is required"))
			return
		}

		// Get the user ID from the header set by middleware
		userID := r.Header.Get(auth.UserIDHeader)
		if userID == "" {
			rest.ErrUnauthorized(w, errors.New("user ID is required"))
			return
		}
		got, err := h.GetHome(r.Context(), &nyum.HomeRequest{
			HomeRequest: nyumpb.HomeRequest{
				HomeId: homeID,
			},
		})
		if err != nil {
			rest.ErrInternal(w, fmt.Errorf("failed to get home: %w", err))
			return
		}
		if got == nil {
			rest.ErrNotFound(w, fmt.Errorf("home with ID %s not found", homeID))
			return
		}
		if got.GetOwnerId() != userID {
			rest.ErrUnauthorized(w, errors.New("user unauthorized to delete this home"))
			return
		}

		resp, err := h.DeleteHome(r.Context(), &nyum.HomeDeleteRequest{
			UserID: userID,
			HomeDeleteRequest: nyumpb.HomeDeleteRequest{
				HomeId: homeID,
			},
		})
		if err != nil {
			if err.Error() == "unauthorized: user does not have permission to delete this home" {
				rest.ErrUnauthorized(w, errors.New("you don't have permission to delete this home"))
				return
			}
			rest.ErrInternal(w, fmt.Errorf("failed to delete home: %w", err))
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
			rest.ErrUnauthorized(w, errors.New("user ID is required"))
			return
		}

		resp, errs := h.GetAllHomesForUser(r.Context(), nyum.UserHomesRequest{
			UserId: userID,
		})
		if len(errs) != 0 {
			errMsgs := make([]string, len(errs))
			for _, err := range errs {
				errMsgs = append(errMsgs, err.Error())
			}
			rest.Mixed(w, http.StatusOK, rest.Result[nyum.HomeResponse]{
				Message: "Some items could not be retrieved",
				Data:    resp,
				Errors:  errMsgs,
			})
			return
		}

		rest.OK(w, resp...)
	}
}
