//go:build integration

package users

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/rkuprov/nyumspace/cmd/serves/internal/sql"
	"github.com/rkuprov/nyumspace/pkg/gen/nyumpb"
	"github.com/rkuprov/nyumspace/pkg/nyum"
	"github.com/rkuprov/nyumspace/pkg/tests"
	"github.com/stretchr/testify/assert"
)

func TestUsers_GetUser(t *testing.T) {
	// initialize test db
	db := tests.DBForTest(t)
	defer tests.CleanupTestDB(t, db)

	users := &Users{
		DB: db, // Replace with actual db connection for real tests
	}

	userID := uuid.NewString()
	ctx := context.Background()
	db.Exec(ctx, sql.RegisterUser,
		userID,
		"testuserNAME",
		"user@nyum.space",
		"hashedpassword", // This should be a bcrypt hash in production
	)

	resp, err := users.GetUser(ctx, &nyum.UserRequest{
		nyumpb.UserRequest{
			UserId: userID,
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, resp.UserId, userID)

	resp, err = users.GetUser(ctx, &nyum.UserRequest{
		nyumpb.UserRequest{
			UserId: "nonexistent-user-id",
		},
	})
	assert.Error(t, err)

	out, err := users.GetUser(ctx, &nyum.UserRequest{
		nyumpb.UserRequest{
			UserId: uuid.NewString(),
		},
	})
	assert.NoError(t, err)
	assert.Nil(t, out, "Expected nil response for non-existent user")
}

func TestUsers_RegisterUser(t *testing.T) {
	// initialize test db
	db := tests.DBForTest(t)
	defer tests.CleanupTestDB(t, db)

	users := &Users{
		DB: db, // Replace with actual db connection for real tests
	}

	ctx := context.Background()
	req := &nyum.UserRegistrationRequest{
		UserRegistrationRequest: nyumpb.UserRegistrationRequest{
			Username: "testuser",
			Email:    "test@nyum.space",
			Password: "password123",
		},
	}

	resp, err := users.RegisterUser(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, resp.GetMessage(), MsgSuccessfulRegistration)

	// Test duplicate registration
	resp, err = users.RegisterUser(ctx, req)
	assert.Error(t, err)
}

func TestUsers_UpdateUser(t *testing.T) {
	// initialize test db
	db := tests.DBForTest(t)
	defer tests.CleanupTestDB(t, db)

	users := &Users{
		DB: db, // Replace with actual db connection for real tests
	}

	userID := uuid.NewString()
	ctx := context.Background()
	db.Exec(ctx, sql.RegisterUser,
		userID,
		"testuser",
		"test@nyum.space",
		"hashedpassword", // This should be a bcrypt hash in production
	)

	req := &nyum.UserUpdateRequest{
		UserID: userID,
		UserUpdateRequest: nyumpb.UserUpdateRequest{
			Username: "updateduser",
			Email:    "updated@nyum.space",
			Password: "newpassword123",
		},
	}
	resp, err := users.UpdateUser(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, resp.Message, MsgSuccessfulUpdate)

	// Verify the update
	updatedResp, err := users.GetUser(ctx, &nyum.UserRequest{
		UserRequest: nyumpb.UserRequest{
			UserId: userID,
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, updatedResp.Username, "updateduser")
	assert.Equal(t, updatedResp.Email, "updated@nyum.space")
}

func TestUsers_DeleteUser(t *testing.T) {
	// initialize test db
	db := tests.DBForTest(t)
	defer tests.CleanupTestDB(t, db)

	users := &Users{
		DB: db, // Replace with actual db connection for real tests
	}

	ctx := context.Background()
	resp, err := users.RegisterUser(ctx, &nyum.UserRegistrationRequest{
		UserRegistrationRequest: nyumpb.UserRegistrationRequest{
			Username: "testuser",
			Email:    "testing@nyum.space",
			Password: "password123",
		},
	})
	userID := resp.GetUserId()
	assert.NoError(t, err)
	deleted, err := users.DeleteUser(ctx, &nyum.UserDeleteRequest{
		nyumpb.UserDeleteRequest{
			UserId: userID,
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, deleted.GetMessage(), MsgSuccessfulDeletion)

	// Verify deletion
	out, err := users.GetUser(ctx, &nyum.UserRequest{
		nyumpb.UserRequest{
			UserId: userID,
		},
	})
	assert.NoError(t, err)
	assert.Nil(t, out)
}
