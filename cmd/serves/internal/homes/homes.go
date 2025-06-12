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
)

type Homes struct {
	DB *pgxpool.Pool
}

func NewHomes(d *daemon.Daemon) *Homes {
	return &Homes{
		DB: d.DB,
	}
}

func (h *Homes) CheckToken(ctx context.Context, token string) (bool, error) {
	var discard string
	var expiresAt time.Time
	err := h.DB.QueryRow(ctx, sql.GetSession, token).Scan(&discard, &expiresAt)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return false, err // Token does not exist
		}
		return false, nil // Token does not exist, but no error
	}

	if expiresAt.Before(time.Now()) {
		return false, nil // Token exists but is expired
	}

	return true, nil
}

// CreateHome creates a new home
func (h *Homes) CreateHome(ctx context.Context, req *nyum.HomeCreationRequest) (*nyum.HomeCreationResponse, error) {
	// Validate required fields
	if req.OwnerId == "" || req.Name == "" {
		return nil, errors.New("owner and name are required")
	}

	homeID := uuid.NewString()

	_, err := h.DB.Exec(ctx, sql.AddHomeSQL,
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
		return nil, fmt.Errorf("home not found: %w", err)
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
	if req.HomeId == "" {
		return nil, errors.New("homeID is required")
	}

	// This is not implemented yet in the handlers, but we'll provide a skeleton
	return &nyum.HomeUpdateResponse{
		HomeUpdateResponse: nyumpb.HomeUpdateResponse{
			Message: "Update home functionality not implemented yet",
		},
	}, nil
}

// DeleteHome deletes a home by ID
func (h *Homes) DeleteHome(ctx context.Context, req *nyum.HomeDeleteRequest) (*nyum.HomeDeleteResponse, error) {
	if req.HomeId == "" {
		return nil, errors.New("homeID is required")
	}

	row := h.DB.QueryRow(ctx, sql.DeleteHomeSQL, req.HomeId)

	var id string
	if err := row.Scan(&id); err != nil {
		return nil, fmt.Errorf("failed to delete home: %w", err)
	}

	return &nyum.HomeDeleteResponse{
		HomeDeleteResponse: nyumpb.HomeDeleteResponse{
			Message: fmt.Sprintf("Home with ID %s deleted successfully", req.HomeId),
		},
	}, nil
}
