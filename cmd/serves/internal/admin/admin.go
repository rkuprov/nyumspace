package admin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rkuprov/nyumspace/cmd/serves/internal/sql"
	"github.com/rkuprov/nyumspace/pkg/daemon"
	"github.com/rkuprov/nyumspace/pkg/gen/nyumpb"
	"github.com/rkuprov/nyumspace/pkg/nyum"
)

type Server interface {
	GetUser(ctx context.Context, req *nyum.UserRequest) (*nyum.UserResponse, error)
	GetAllUsers(ctx context.Context) ([]*nyum.UserResponse, error)
	DeleteUser(ctx context.Context, req *nyum.UserDeleteRequest) (*nyum.UserDeleteResponse, error)
	GetHome(ctx context.Context, req *nyum.HomeRequest) (*nyum.HomeResponse, error)
	GetAllHomes(ctx context.Context) ([]*nyum.HomeResponse, error)
	DeleteHome(ctx context.Context, req *nyum.HomeDeleteRequest) (*nyum.HomeDeleteResponse, error)
	GetHomesForUser(ctx context.Context, req *nyum.UserHomesRequest) ([]*nyum.HomeResponse, error)
}

var _ Server = (*Admin)(nil)

type Admin struct {
	DB *pgxpool.Pool
}

func NewAdmin(d daemon.Daemon) Admin {
	return Admin{
		DB: d.DB,
	}
}

func (a *Admin) GetAllUsers(ctx context.Context) ([]*nyum.UserResponse, error) {
	rows, err := a.DB.Query(ctx, sql.GetAllUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	users := make([]*nyum.UserResponse, 0)
	for rows.Next() {
		var user nyum.UserResponse
		if err := rows.Scan(&user.UserId, &user.Username, &user.Email); err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}

		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over users: %w", err)
	}

	return users, nil
}

// GetAllHomes retrieves all homes
func (a *Admin) GetAllHomes(ctx context.Context) ([]*nyum.HomeResponse, error) {
	rows, err := a.DB.Query(ctx, sql.GetAllHomesSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to get homes: %w", err)
	}
	defer rows.Close()

	homes := make([]*nyum.HomeResponse, 0)
	for rows.Next() {
		var home nyum.HomeResponse
		var createdAt, updatedAt time.Time
		var description, imageURL pgtype.Text

		if err := rows.Scan(
			&home.HomeId,
			&home.OwnerId,
			&home.Name,
			&home.StreetAddress_1,
			&home.StreetAddress_2,
			&home.City,
			&home.State,
			&home.ZipCode,
			&home.Country,
			&description,
			&home.Tags,
			&imageURL,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan home row: %w", err)
		}

		home.CreatedAt = createdAt.Format(time.RFC3339)
		home.UpdatedAt = updatedAt.Format(time.RFC3339)

		home.Description = description.String
		home.ImageUrl = imageURL.String
		homes = append(homes, &home)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over homes: %w", err)
	}

	return homes, nil
}

func (a *Admin) DeleteUser(ctx context.Context, req *nyum.UserDeleteRequest) (*nyum.UserDeleteResponse, error) {
	if req.UserId == "" {
		return nil, errors.New("userID is required")
	}

	_, err := a.DB.Exec(ctx, sql.DeleteUser, req.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to delete user: %w", err)
	}

	return &nyum.UserDeleteResponse{
		UserDeleteResponse: nyumpb.UserDeleteResponse{
			Message: fmt.Sprintf("User with ID %s deleted successfully", req.UserId),
		},
	}, nil
}

// DeleteHome deletes a home by ID
func (a *Admin) DeleteHome(ctx context.Context, req *nyum.HomeDeleteRequest) (*nyum.HomeDeleteResponse, error) {
	if req.HomeId == "" {
		return nil, errors.New("homeID is required")
	}

	row := a.DB.QueryRow(ctx, sql.DeleteHomeSQL, req.HomeId)

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

// GetHome retrieves a home by ID
func (a *Admin) GetHome(ctx context.Context, req *nyum.HomeRequest) (*nyum.HomeResponse, error) {
	if req.HomeId == "" {
		return nil, errors.New("homeID is required")
	}

	row := a.DB.QueryRow(ctx, sql.GetHomeSQL, req.HomeId)

	var home nyum.HomeResponse
	var createdAt, updatedAt time.Time

	if err := row.Scan(
		&home.HomeId,
		&home.OwnerId,
		&home.Name,
		&home.StreetAddress_1,
		&home.StreetAddress_2,
		&home.City,
		&home.State,
		&home.ZipCode,
		&home.Country,
		&home.Description,
		&home.Tags,
		&home.ImageUrl,
		&createdAt,
		&updatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("home not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get home: %w", err)
	}

	home.CreatedAt = createdAt.Format(time.RFC3339)
	home.UpdatedAt = updatedAt.Format(time.RFC3339)

	return &home, nil
}

// GetHomesForUser retrieves all homes for a specific user
func (a *Admin) GetHomesForUser(ctx context.Context, req *nyum.UserHomesRequest) ([]*nyum.HomeResponse, error) {
	if req.UserId == "" {
		return nil, errors.New("userID is required")
	}

	rows, err := a.DB.Query(ctx, sql.GetAllHomesForUserSQL, req.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to get homes for user: %w", err)
	}
	defer rows.Close()

	homes := make([]*nyum.HomeResponse, 0)
	for rows.Next() {
		var home nyum.HomeResponse
		var createdAt, updatedAt time.Time

		if err := rows.Scan(
			&home.HomeId,
			&home.OwnerId,
			&home.Name,
			&home.StreetAddress_1,
			&home.StreetAddress_2,
			&home.City,
			&home.State,
			&home.ZipCode,
			&home.Country,
			&home.Description,
			&home.Tags,
			&home.ImageUrl,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan home row: %w", err)
		}

		home.CreatedAt = createdAt.Format(time.RFC3339)
		home.UpdatedAt = updatedAt.Format(time.RFC3339)

		homes = append(homes, &home)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over homes: %w", err)
	}

	return homes, nil
}

// GetUser retrieves a user by ID
func (a *Admin) GetUser(ctx context.Context, req *nyum.UserRequest) (*nyum.UserResponse, error) {
	if req.UserId == "" {
		return nil, errors.New("userID is required")
	}

	row := a.DB.QueryRow(ctx, sql.GetUser, req.UserId)

	var user nyum.UserResponse
	if err := row.Scan(&user.UserId, &user.Username, &user.Email); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}
