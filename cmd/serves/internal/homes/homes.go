package homes

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rkuprov/nyumspace/cmd/serves/internal/sql"
	"github.com/rkuprov/nyumspace/pkg/daemon"
	"github.com/rkuprov/nyumspace/pkg/gen/nyumpb"
	"github.com/rkuprov/nyumspace/pkg/nyum"
	"github.com/rkuprov/nyumspace/pkg/store"
)

var (
	ErrNotFound = errors.New("not found")
)

type Homes struct {
	DB    *pgxpool.Pool
	Store store.Store
}

func NewHomes(d *daemon.Daemon, s store.Store) *Homes {
	return &Homes{
		DB:    d.DB,
		Store: s,
	}
}

// CreateHome creates a new home
func (h *Homes) CreateHome(ctx context.Context, req *nyum.HomeCreationRequest) (*nyum.HomeCreationResponse, error) {
	// Validate required fields
	if req.UserID == "" {
		return nil, errors.New("owner is required")
	}

	// Set default name if empty
	if req.Name == "" {
		req.Name = "Unnamed Home"
	}

	homeID := uuid.NewString()

	_, err := h.DB.Exec(ctx, sql.AddHomeSQL,
		homeID,
		req.UserID,
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
		return nil, fmt.Errorf("failed to create home: %w", err)
	}

	return &nyum.HomeCreationResponse{
		HomeCreationResponse: nyumpb.HomeCreationResponse{
			HomeId:  homeID,
			Message: fmt.Sprintf("Home '%s' created successfully", req.Name),
		},
	}, nil
}

// GetHome retrieves a home by ID
func (h *Homes) GetHome(ctx context.Context, req *nyum.HomeRequest) (*nyum.HomeResponse, error) {
	if req.HomeId == "" {
		return nil, errors.New("homeID is required")
	}

	row := h.DB.QueryRow(ctx, sql.GetHomeSQL, req.HomeId)

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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &nyum.HomeResponse{
		HomeResponse: nyumpb.HomeResponse{
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
		},
	}, nil
}

// UpdateHome updates a home (not yet implemented)
func (h *Homes) UpdateHome(ctx context.Context, req *nyum.HomeUpdateRequest) (*nyum.HomeUpdateResponse, error) {
	_, err := h.DB.Exec(ctx, sql.UpdateHomeSQL,
		req.HomeID,
		req.GetName(),
		req.GetDescription(),
		req.GetStreetAddress_1(),
		req.GetStreetAddress_2(),
		req.GetCity(),
		req.GetState(),
		req.GetZipCode(),
		req.GetCountry(),
		req.GetImageUrl(),
		req.GetTags(),
		time.Now().Format(time.RFC3339),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update home: %w", err)
	}

	// This is not implemented yet in the handlers, but we'll provide a skeleton
	return &nyum.HomeUpdateResponse{
		HomeUpdateResponse: nyumpb.HomeUpdateResponse{
			Message: fmt.Sprintf("Home '%s' updated successfully", req.HomeID),
		},
	}, nil
}

// DeleteHome deletes a home by ID and its associated images
func (h *Homes) DeleteHome(ctx context.Context, req *nyum.HomeDeleteRequest) (*nyum.HomeDeleteResponse, error) {
	// First, get the home to check for images to delete
	home, err := h.GetHome(ctx, &nyum.HomeRequest{
		HomeRequest: nyumpb.HomeRequest{
			HomeId: req.HomeId,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get home for deletion: %w", err)
	}
	if home == nil {
		return nil, ErrNotFound
	}

	// Delete the home from database
	var id string
	err = h.DB.QueryRow(ctx, sql.DeleteHomeSQL, req.HomeId).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to delete home: %w", err)
	}

	// Delete associated image from storage if it exists
	if home.ImageUrl != "" && h.Store != nil {
		if s3Store, ok := h.Store.(*store.S3Store); ok {
			if key, err := s3Store.ExtractKeyFromURL(home.ImageUrl); err == nil {
				// Best effort - don't fail the deletion if image deletion fails
				_ = h.Store.Delete(ctx, key)
			}
		}
	}

	return &nyum.HomeDeleteResponse{
		HomeDeleteResponse: nyumpb.HomeDeleteResponse{
			Message: fmt.Sprintf("Home with ID %s deleted successfully", req.HomeId),
		},
	}, nil
}

// GetAllHomesForUser retrieves all homes belonging to a specific user
func (h *Homes) GetAllHomesForUser(ctx context.Context, req nyum.UserHomesRequest) ([]nyum.HomeResponse, []error) {
	rows, err := h.DB.Query(ctx, sql.GetAllHomesForUserSQL, req.UserId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, []error{fmt.Errorf("failed to get homes for user: %w", err)}
	}
	defer rows.Close()

	homes := []nyum.HomeResponse{}
	var errList []error
	for rows.Next() {
		// todo: rework to use a struct for scanning
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

		if err := rows.Scan(
			&id,
			&ownerID,
			&name,
			&streetAddress1,
			&streetAddress2,
			&city,
			&state,
			&zipCode,
			&country,
			&description,
			&tags,
			&imageURL,
			&createdAt,
			&updatedAt,
		); err != nil {
			errList = append(errList, fmt.Errorf("failed to scan home row: %w", err))
		}

		homes = append(homes, nyum.HomeResponse{
			HomeResponse: nyumpb.HomeResponse{
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
			},
		})
	}

	return homes, errList
}
