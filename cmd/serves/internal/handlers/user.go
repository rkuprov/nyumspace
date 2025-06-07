package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/rkuprov/nyumspace/cmd/serves/internal/sql"
	"golang.org/x/crypto/bcrypt"

	"connectrpc.com/connect"

	"github.com/rkuprov/nyumspace/pkg/gen/nyumpb"
)

// RegisterUser registers a new user in the system
func (s *ServerHandler) RegisterUser(
	ctx context.Context,
	req *connect.Request[nyumpb.UserRegistrationRequest],
) (*connect.Response[nyumpb.UserRegistrationResponse], error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Msg.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	id := uuid.NewString()
	row := s.db.QueryRow(ctx, sql.RegisterUser, id, req.Msg.GetUsername(), req.Msg.GetEmail(), string(hash))
	if err = row.Scan(&id); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	log.Printf(
		"Registered user %s with email %s with hash %s", id, req.Msg.GetEmail(), string(hash),
	)

	return &connect.Response[nyumpb.UserRegistrationResponse]{
		Msg: &nyumpb.UserRegistrationResponse{
			Success: true,
			Message: fmt.Sprintf("User %s registered successfully with ID: %s", req.Msg.GetUsername(), id),
			UserId:  id,
		},
	}, nil
}

func (s *ServerHandler) GetUser(ctx context.Context, req *connect.Request[nyumpb.UserRequest]) (*connect.Response[nyumpb.UserResponse], error) {
	row := s.db.QueryRow(ctx, sql.GetUser, req.Msg.GetUserId())
	var id, name, email string
	if err := row.Scan(&id, &name, &email); err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user not found: %w", err))
	}
	return &connect.Response[nyumpb.UserResponse]{
		Msg: &nyumpb.UserResponse{
			UserId:   id,
			Username: name,
			Email:    email,
		},
	}, nil
}

func (s *ServerHandler) UpdateUser(ctx context.Context, req *connect.Request[nyumpb.UserUpdateRequest]) (*connect.Response[nyumpb.UserUpdateResponse], error) {
	row := s.db.QueryRow(ctx, sql.UpdateUser, req.Msg.GetUserId(), req.Msg.GetUsername(), req.Msg.GetEmail())
	var id int32
	if err := row.Scan(&id); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update user: %w", err))
	}
	return &connect.Response[nyumpb.UserUpdateResponse]{
		Msg: &nyumpb.UserUpdateResponse{
			Success: true,
			Message: fmt.Sprintf("User with ID %d updated successfully", id),
		},
	}, nil
}

func (s *ServerHandler) DeleteUser(ctx context.Context, req *connect.Request[nyumpb.UserDeleteRequest]) (*connect.Response[nyumpb.UserDeleteResponse], error) {
	row := s.db.QueryRow(ctx, sql.DeleteUser, req.Msg.GetUserId())
	var id int32
	if err := row.Scan(&id); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to delete user: %w", err))
	}
	return &connect.Response[nyumpb.UserDeleteResponse]{
		Msg: &nyumpb.UserDeleteResponse{
			Success: true,
			Message: fmt.Sprintf("User with ID %d deleted successfully", id),
		},
	}, nil
}

func (s *ServerHandler) LoginUser(ctx context.Context, req *connect.Request[nyumpb.UserLoginRequest]) (*connect.Response[nyumpb.UserLoginResponse], error) {
	var userID, hashedPassword string

	row := s.db.QueryRow(ctx, sql.GetUserByEmail, req.Msg.GetEmail())
	if err := row.Scan(&userID, &hashedPassword); err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user not found: %w", err))
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Msg.GetPassword())); err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid password: %w", err))
	}

	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to generate session token: %w", err))
	}
	sessionToken := hex.EncodeToString(tokenBytes)

	expiresAt := time.Now().Add(24 * time.Hour)

	_, err := s.db.Exec(ctx, sql.CreateSession, sessionToken, userID, expiresAt)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create session: %w", err))
	}

	return &connect.Response[nyumpb.UserLoginResponse]{
		Msg: &nyumpb.UserLoginResponse{
			Success:      true,
			Message:      "Login successful",
			SessionToken: sessionToken,
			UserId:       fmt.Sprintf("%s", userID),
		},
	}, nil
}

func (s *ServerHandler) LogoutUser(ctx context.Context, req *connect.Request[nyumpb.UserLogoutRequest]) (*connect.Response[nyumpb.UserLogoutResponse], error) {
	_, err := s.db.Exec(ctx, sql.DeleteSession, req.Msg.GetSessionToken())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to delete session: %w", err))
	}

	return &connect.Response[nyumpb.UserLogoutResponse]{
		Msg: &nyumpb.UserLogoutResponse{
			Success: true,
			Message: "Logout successful",
		},
	}, nil
}
