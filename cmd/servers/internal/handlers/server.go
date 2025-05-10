package handlers

import (
	"github.com/rkuprov/nyumspace/pkg/gen/nyumpb/nyumpbconnect"
)

// Ensure ServerHandler implements the ServerService interface
var _ nyumpbconnect.ServerServiceHandler = (*ServerHandler)(nil)

type ServerHandler struct {
	// Add any dependencies you need here, like:
	// userRepo repositories.UserRepository
	// homeRepo repositories.HomeRepository
	// etc.
}

// NewServerHandler creates a new instance of ServerHandler
func NewServerHandler() *ServerHandler {
	return &ServerHandler{
		// Initialize your dependencies here
	}
}
