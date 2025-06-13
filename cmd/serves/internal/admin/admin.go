package admin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rkuprov/nyumspace/cmd/serves/internal/sql"
	"github.com/rkuprov/nyumspace/pkg/daemon"
	"github.com/rkuprov/nyumspace/pkg/gen/nyumpb"
	"github.com/rkuprov/nyumspace/pkg/nyum"
)

type Admin struct {
	DB *pgxpool.Pool
}

func NewAdmin(d daemon.Daemon) Admin {
	return Admin{
		DB: d.DB,
	}
}

func (a *Admin) GetAllUsers(ctx context.Context) ([]nyum.UserResponse, error) {
	rows, err := a.DB.Query(ctx, sql.GetAllUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	users := []nyum.UserResponse{}
	for rows.Next() {
		var user nyum.UserResponse
		if err := rows.Scan(&user.UserId, &user.Username, &user.Email); err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over users: %w", err)
	}

	return users, nil
}

// GetAllHomes retrieves all homes
func (a *Admin) GetAllHomes(ctx context.Context) ([]nyum.HomeResponse, error) {
	rows, err := a.DB.Query(ctx, sql.GetAllHomesSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to get homes: %w", err)
	}
	defer rows.Close()

	homes := []nyum.HomeResponse{}
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

		homes = append(homes, home)
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
