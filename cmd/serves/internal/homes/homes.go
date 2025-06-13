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

var ErrUserUnauthorized = errors.New("unauthorized: user does not have permission to access this resource")

type Homes struct {
	DB *pgxpool.Pool
}

func NewHomes(d *daemon.Daemon) *Homes {
	return &Homes{
		DB: d.DB,
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
		&zipCode, &country, &description, &tags, &imageURL, &createdAt, &updatedAt); err != nil && errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	// Authorization check: Ensure the requesting user is the owner of the home
	if req.GetOwnerId() != "" && ownerID != req.GetOwnerId() {
		return nil, ErrUserUnauthorized
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
	if req.HomeID == "" {
		return nil, errors.New("homeID is required")
	}
	if req.UserID == "" {
		return nil, errors.New("userID is required")
	}
	// First, verify that the home exists and the requesting user is the owner
	var ownerID string
	err := h.DB.QueryRow(ctx, "SELECT owner_id FROM homes WHERE id = $1", req.HomeID).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("home not found: %w", err)
		}
		return nil, fmt.Errorf("failed to retrieve home owner: %w", err)
	}

	_, err = h.DB.Exec(ctx, sql.UpdateHomeSQL,
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

// DeleteHome deletes a home by ID
func (h *Homes) DeleteHome(ctx context.Context, req *nyum.HomeDeleteRequest) (*nyum.HomeDeleteResponse, error) {
	if req.HomeId == "" {
		return nil, errors.New("homeID is required")
	}

	// First, verify that the home exists and the requesting user is the owner
	if req.UserID == "" {
		return nil, errors.New("userID is required")
	}

	row := h.DB.QueryRow(ctx, "SELECT owner_id FROM homes WHERE id = $1", req.HomeId)
	var ownerID string
	if err := row.Scan(&ownerID); err != nil {
		return nil, fmt.Errorf("home not found: %w", err)
	}

	// Authorization check: Ensure the requesting user is the owner of the home
	if ownerID != req.UserID {
		return nil, fmt.Errorf("unauthorized: user does not have permission to delete this home")
	}

	// If authorization passes or isn't required, proceed with deletion
	var id string
	err := h.DB.QueryRow(ctx, sql.DeleteHomeSQL, req.HomeId).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to delete home: %w", err)
	}

	return &nyum.HomeDeleteResponse{
		HomeDeleteResponse: nyumpb.HomeDeleteResponse{
			Message: fmt.Sprintf("Home with ID %s deleted successfully", req.HomeId),
		},
	}, nil
}

// GetAllHomesForUser retrieves all homes belonging to a specific user
func (h *Homes) GetAllHomesForUser(ctx context.Context, userID string) ([]nyum.HomeResponse, error) {
	if userID == "" {
		return nil, errors.New("userID is required")
	}

	rows, err := h.DB.Query(ctx, sql.GetAllHomesForUserSQL, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get homes for user: %w", err)
	}
	defer rows.Close()

	homes := []nyum.HomeResponse{}
	for rows.Next() {
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
			return nil, fmt.Errorf("failed to scan home row: %w", err)
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

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over homes: %w", err)
	}

	return homes, nil
}
