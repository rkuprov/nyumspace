package handlers

import (
	"context"

	"connectrpc.com/connect"

	"github.com/rkuprov/nyumspace/pkg/gen/nyumpb"
)

// RegisterUser(context.Context, *connect.Request[nyumpb.UserRegistrationRequest]) (*connect.Response[nyumpb.UserRegistrationResponse], error)

// User-related methods
func (s *ServerHandler) RegisterUser(ctx context.Context, req *connect.Request[nyumpb.UserRegistrationRequest]) (*connect.Response[nyumpb.UserRegistrationResponse], error) {
	// Implementation goes here
	return &connect.Response[nyumpb.UserRegistrationResponse]{}, nil
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
