package handlers

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/rkuprov/nyumspace/cmd/serves/internal/sql"
	"github.com/rkuprov/nyumspace/pkg/gen/nyumpb"
)

// Home-related methods
func (s *ServerHandler) AddHome(ctx context.Context, req *connect.Request[nyumpb.HomeCreationRequest]) (*connect.Response[nyumpb.HomeCreationResponse], error) {
	homeID := uuid.New().String()

	_, err := s.db.Exec(ctx, sql.AddHomeSQL,
		homeID,
		req.Msg.OwnerId,
		req.Msg.Name,
		req.Msg.StreetAddress_1,
		req.Msg.StreetAddress_2,
		req.Msg.City,
		req.Msg.State,
		req.Msg.ZipCode,
		req.Msg.Country,
		req.Msg.Description,
		req.Msg.Tags,
		req.Msg.ImageUrl,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create home: %w", err))
	}

	return &connect.Response[nyumpb.HomeCreationResponse]{
		Msg: &nyumpb.HomeCreationResponse{
			HomeId:  homeID,
			Success: true,
			Message: fmt.Sprintf("Home '%s' created successfully", req.Msg.Name),
		},
	}, nil
}

func (s *ServerHandler) GetHome(ctx context.Context, req *connect.Request[nyumpb.HomeRequest]) (*connect.Response[nyumpb.HomeResponse], error) {
	row := s.db.QueryRow(ctx, sql.GetHomeSQL, req.Msg.GetHomeId())

	var (
		homeID         string
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

	if err := row.Scan(&homeID, &ownerID, &name, &streetAddress1, &streetAddress2, &city, &state,
		&zipCode, &country, &description, &tags, &imageURL, &createdAt, &updatedAt); err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("home not found: %w", err))
	}

	return &connect.Response[nyumpb.HomeResponse]{
		Msg: &nyumpb.HomeResponse{
			HomeId:          homeID,
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

func (s *ServerHandler) UpdateHome(ctx context.Context, req *connect.Request[nyumpb.HomeUpdateRequest]) (*connect.Response[nyumpb.HomeUpdateResponse], error) {
	return nil, nil
}

func (s *ServerHandler) DeleteHome(ctx context.Context, req *connect.Request[nyumpb.HomeDeleteRequest]) (*connect.Response[nyumpb.HomeDeleteResponse], error) {
	row := s.db.QueryRow(ctx, sql.DeleteHomeSQL, req.Msg.HomeId)

	var homeID string
	if err := row.Scan(&homeID); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to delete home: %w", err))
	}

	return &connect.Response[nyumpb.HomeDeleteResponse]{
		Msg: &nyumpb.HomeDeleteResponse{
			Success: true,
			Message: fmt.Sprintf("Home with ID %s deleted successfully", homeID),
		},
	}, nil
}
