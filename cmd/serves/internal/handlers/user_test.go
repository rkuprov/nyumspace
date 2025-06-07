package handlers

import (
	"context"
	"fmt"
	"log"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rkuprov/nyumspace/pkg/daemon"
	"github.com/rkuprov/nyumspace/pkg/gen/nyumpb"
	"github.com/rkuprov/nyumspace/pkg/tests"
)

func TestServerHandler_RegisterUser(t *testing.T) {
	pool := tests.DBForTest(t)
	dbname := pool.Config().ConnConfig.Database

	defer func() {
		err := tests.RemoveDBForTest(dbname)
		if err != nil {
			log.Fatalf("failed to remove test database: %v", err)
		}
	}()
	defer func() {
		pool.Close()
	}()

	svs := NewServerHandler(daemon.Daemon{
		DB:     pool,
		Server: nil,
		Router: nil,
	})
	req := &connect.Request[nyumpb.UserRegistrationRequest]{
		Msg: &nyumpb.UserRegistrationRequest{
			Username: "testuser",
			Email:    "test@test.com",
			Password: "testpassword",
		},
	}

	resp, err := svs.RegisterUser(context.Background(), req)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}
	assert.True(t, resp.Msg.GetSuccess())
	id := resp.Msg.GetUserId()

	// Verify in DB
	var userID string
	err = pool.QueryRow(context.Background(), "SELECT id FROM users WHERE name = $1", req.Msg.GetUsername()).Scan(&userID)
	assert.NoError(t, err)
	assert.Equal(t, id, userID)
}

func TestServerHandler_GetUser(t *testing.T) {
	pool := tests.DBForTest(t)
	dbname := pool.Config().ConnConfig.Database

	defer func() {
		err := tests.RemoveDBForTest(dbname)
		if err != nil {
			log.Fatalf("failed to remove test database: %v", err)
		}
	}()
	defer pool.Close()

	svs := NewServerHandler(daemon.Daemon{
		DB: pool,
	})

	// Register a test user
	registerReq := &connect.Request[nyumpb.UserRegistrationRequest]{
		Msg: &nyumpb.UserRegistrationRequest{
			Username: "testuser",
			Email:    "test@test.com",
			Password: "testpassword",
		},
	}
	registerResp, err := svs.RegisterUser(context.Background(), registerReq)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}
	assert.True(t, registerResp.Msg.GetSuccess())
	userID := registerResp.Msg.GetUserId()

	// Retrieve the user
	getReq := &connect.Request[nyumpb.UserRequest]{
		Msg: &nyumpb.UserRequest{
			UserId: userID,
		},
	}
	getResp, err := svs.GetUser(context.Background(), getReq)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}

	// Verify the response
	assert.Equal(t, userID, getResp.Msg.GetUserId())
	assert.Equal(t, "testuser", getResp.Msg.GetUsername())
	assert.Equal(t, "test@test.com", getResp.Msg.GetEmail())
}

func TestServerHandler_DeleteUser(t *testing.T) {
	pool := tests.DBForTest(t)
	dbname := pool.Config().ConnConfig.Database

	defer func() {
		err := tests.RemoveDBForTest(dbname)
		if err != nil {
			log.Fatalf("failed to remove test database: %v", err)
		}
	}()
	defer pool.Close()

	svs := NewServerHandler(daemon.Daemon{
		DB: pool,
	})

	// Register a test user
	registerReq := &connect.Request[nyumpb.UserRegistrationRequest]{
		Msg: &nyumpb.UserRegistrationRequest{
			Username: "testuser",
			Email:    "test@test.com",
			Password: "testpassword",
		},
	}
	registerResp, err := svs.RegisterUser(context.Background(), registerReq)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}
	assert.True(t, registerResp.Msg.GetSuccess())
	userID := registerResp.Msg.GetUserId()

	// Delete the user
	deleteReq := &connect.Request[nyumpb.UserDeleteRequest]{
		Msg: &nyumpb.UserDeleteRequest{
			UserId: userID,
		},
	}
	deleteResp, err := svs.DeleteUser(context.Background(), deleteReq)
	require.NoError(t, err)

	// Verify the response
	assert.True(t, deleteResp.Msg.GetSuccess())
	assert.Equal(t, fmt.Sprintf("User with ID %s deleted successfully", userID), deleteResp.Msg.GetMessage())

	// Ensure the user no longer exists in the database
	var count int
	err = pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM users WHERE id = $1", userID).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count, "User was not deleted from the database")
}
