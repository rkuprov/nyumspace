package handlers

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rkuprov/nyumspace/pkg/daemon"
	"github.com/rkuprov/nyumspace/pkg/gen/nyumpb/nyumpbconnect"
)

// Ensure ServerHandler implements the ServerService interface
var _ nyumpbconnect.ServerServiceHandler = (*ServerHandler)(nil)

type ServerHandler struct {
	db *pgxpool.Pool // Replace with your actual DB type
	// Add any dependencies you need here, like:
	// userRepo repositories.UserRepository
	// homeRepo repositories.HomeRepository
	// etc.
}

// NewServerHandler creates a new instance of ServerHandler
func NewServerHandler(d daemon.Daemon) *ServerHandler {
	return &ServerHandler{
		db: d.DB, // Initialize your dependencies here
	}
}
