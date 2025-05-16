package handlers

import (
	"context"
	"fmt"
	"github.com/rkuprov/nyumspace/cmd/serves/internal/sql"
	"golang.org/x/crypto/bcrypt"

	"connectrpc.com/connect"

	"github.com/rkuprov/nyumspace/pkg/gen/nyumpb"
)

// RegisterUser(context.Context, *connect.Request[nyumpb.UserRegistrationRequest]) (*connect.Response[nyumpb.UserRegistrationResponse], error)

// User-related methods

// RegisterUser registers a new user in the system
func (s *ServerHandler) RegisterUser(
	ctx context.Context,
	req *connect.Request[nyumpb.UserRegistrationRequest],
) (*connect.Response[nyumpb.UserRegistrationResponse], error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Msg.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	row := s.db.QueryRow(ctx, sql.RegisterUser, req.Msg.GetUsername(), req.Msg.GetEmail(), string(hash))
	var id int
	if err = row.Scan(&id); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return &connect.Response[nyumpb.UserRegistrationResponse]{
		Msg: &nyumpb.UserRegistrationResponse{
			Success: true,
			Message: fmt.Sprintf("User %s registered successfully with ID: %d", req.Msg.GetUsername(), id),
		},
	}, nil
}

func (s *ServerHandler) GetUser(ctx context.Context, req *connect.Request[nyumpb.UserRequest]) (*connect.Response[nyumpb.UserResponse], error) {
	// Implementation goes here
	return &connect.Response[nyumpb.UserResponse]{}, nil
}

func (s *ServerHandler) UpdateUser(ctx context.Context, req *connect.Request[nyumpb.UserUpdateRequest]) (*connect.Response[nyumpb.UserUpdateResponse], error) {
	// Implementation goes here
	return &connect.Response[nyumpb.UserUpdateResponse]{}, nil
}

func (s *ServerHandler) DeleteUser(ctx context.Context, req *connect.Request[nyumpb.UserDeleteRequest]) (*connect.Response[nyumpb.UserDeleteResponse], error) {
	// Implementation goes here
	return &connect.Response[nyumpb.UserDeleteResponse]{}, nil
}
