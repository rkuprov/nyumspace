package handlers

import (
	"context"
	"log"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"

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
	assert.Equal(t, "User testuser registered successfully with ID: 1", resp.Msg.GetMessage())
}
