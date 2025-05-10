package handlers

import (
	"context"

	"connectrpc.com/connect"

	"github.com/rkuprov/nyumspace/pkg/gen/nyumpb"
)

// Home-related methods
func (s *ServerHandler) AddHome(ctx context.Context, req *connect.Request[nyumpb.HomeCreationRequest]) (*connect.Response[nyumpb.HomeCreationResponse], error) {
	// Implementation goes here
	return &connect.Response[nyumpb.HomeCreationResponse]{}, nil
}

func (s *ServerHandler) GetHome(ctx context.Context, req *connect.Request[nyumpb.HomeRequest]) (*connect.Response[nyumpb.HomeResponse], error) {
	// Implementation goes here
	return &connect.Response[nyumpb.HomeResponse]{}, nil
}

func (s *ServerHandler) UpdateHome(ctx context.Context, req *connect.Request[nyumpb.HomeUpdateRequest]) (*connect.Response[nyumpb.HomeUpdateResponse], error) {
	// Implementation goes here
	return &connect.Response[nyumpb.HomeUpdateResponse]{}, nil
}

func (s *ServerHandler) DeleteHome(ctx context.Context, req *connect.Request[nyumpb.HomeDeleteRequest]) (*connect.Response[nyumpb.HomeDeleteResponse], error) {
	// Implementation goes here
	return &connect.Response[nyumpb.HomeDeleteResponse]{}, nil
}
