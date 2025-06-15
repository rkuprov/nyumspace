package users

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rkuprov/nyumspace/cmd/serves/internal/sql"
	"github.com/rkuprov/nyumspace/pkg/daemon"
	"github.com/rkuprov/nyumspace/pkg/gen/nyumpb"
	"github.com/rkuprov/nyumspace/pkg/nyum"
	"golang.org/x/crypto/bcrypt"
)

var ErrNotFound = errors.New("not found")

type Users struct {
	DB *pgxpool.Pool
}

func NewUsers(d *daemon.Daemon) *Users {
	return &Users{
		DB: d.DB,
	}
}

func (u *Users) RegisterUser(ctx context.Context, req *nyum.UserRegistrationRequest) (*nyum.UserRegistrationResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	id := uuid.NewString()
	row := u.DB.QueryRow(ctx, sql.RegisterUser, id, req.Username, req.Email, string(hash))
	// todo: implement response for failed registration due to duplicate email or username
	if err = row.Scan(&id); err != nil {
		return nil, err
	}

	return &nyum.UserRegistrationResponse{
		UserRegistrationResponse: nyumpb.UserRegistrationResponse{
			UserId:  id,
			Message: "User registered successfully",
		},
	}, nil
}

// GetUser retrieves a user by their ID. User ID is required and must be provided.
func (u *Users) GetUser(ctx context.Context, req *nyum.UserRequest) (*nyum.UserResponse, error) {
	row := u.DB.QueryRow(ctx, sql.GetUser, req.UserId)
	var id, name, email string
	if err := row.Scan(&id, &name, &email); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &nyum.UserResponse{
		UserResponse: nyumpb.UserResponse{
			UserId:   id,
			Username: name,
			Email:    email,
		},
	}, nil
}

func (u *Users) UpdateUser(ctx context.Context, req *nyum.UserUpdateRequest) (*nyum.UserUpdateResponse, error) {
	var id string
	var hashedPassword []byte
	var err error
	if req.GetPassword() != "" {
		hashedPassword, err = bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
	}

	err = u.DB.QueryRow(ctx, sql.UpdateUser, req.UserId, req.Username, req.Email, string(hashedPassword)).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &nyum.UserUpdateResponse{
		UserUpdateResponse: nyumpb.UserUpdateResponse{
			Message: fmt.Sprintf("User with ID %s updated successfully", req.UserId),
		},
	}, nil
}

func (u *Users) DeleteUser(ctx context.Context, req *nyum.UserDeleteRequest) (*nyum.UserDeleteResponse, error) {
	if req.UserId == "" {
		return nil, errors.New("userID is required")
	}

	_, err := u.DB.Exec(ctx, sql.DeleteUser, req.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to delete user: %w", err)
	}

	return &nyum.UserDeleteResponse{
		UserDeleteResponse: nyumpb.UserDeleteResponse{
			Message: fmt.Sprintf("User with ID %s deleted successfully", req.UserId),
		},
	}, nil
}

func (u *Users) LoginUser(ctx context.Context, req *nyum.UserLoginRequest) (*nyum.UserLoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("email and password are required")
	}

	var userID, hashedPassword string
	row := u.DB.QueryRow(ctx, sql.GetUserByEmail, req.Email)
	if err := row.Scan(&userID, &hashedPassword); err != nil {
		return nil, errors.New("could not find user with provided email")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		log.Println(err)
		return nil, errors.New("invalid email or password")
	}

	// Generate session token
	sessionToken := uuid.NewString()
	expiresAt := time.Now().Add(24 * time.Hour)
	_, err := u.DB.Exec(ctx, sql.CreateSession, sessionToken, userID, expiresAt)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &nyum.UserLoginResponse{
		UserLoginResponse: nyumpb.UserLoginResponse{
			Message:      "Login successful",
			SessionToken: sessionToken,
			UserId:       userID,
		},
	}, nil
}

func (u *Users) LogoutUser(ctx context.Context, req *nyum.UserLogoutRequest) (*nyum.UserLogoutResponse, error) {
	if req.SessionToken == "" {
		return nil, errors.New("session token is required")
	}

	_, err := u.DB.Exec(ctx, sql.DeleteSession, req.SessionToken)
	if err != nil {
		return nil, fmt.Errorf("failed to delete session: %w", err)
	}

	return &nyum.UserLogoutResponse{
		UserLogoutResponse: nyumpb.UserLogoutResponse{
			Message: "Logout successful",
		},
	}, nil
}
