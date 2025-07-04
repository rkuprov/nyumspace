package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	check "github.com/rkuprov/checkpoint"
	"github.com/rkuprov/nyumspace/cmd/serves/internal/users"
	"github.com/rkuprov/nyumspace/pkg/auth"
	"github.com/rkuprov/nyumspace/pkg/daemon"
	"github.com/rkuprov/nyumspace/pkg/gen/nyumpb"
	"github.com/rkuprov/nyumspace/pkg/nyum"
	"github.com/rkuprov/nyumspace/pkg/rest"
	"github.com/rkuprov/nyumspace/pkg/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserEndpoints(t *testing.T) {
	ctx := context.Background()
	db := tests.DBForTest(t)
	defer tests.CleanupTestDB(t, db)

	d := daemon.Daemon{DB: db}
	u := users.NewUsers(&d)

	checkFunc := check.NewChecker(chi.NewRouter())

	// Create a test user
	resp, err := checkFunc(
		ctx,
		http.HandlerFunc(RegisterUser(u)),
		check.WithURLPath("/register"),
		check.WithURLPattern("/register"),
		check.WithBody(`{
			"username": "testuser2", 
			"email": "testuser@nyum.space", 
			"password": "testpassword"}`,
		),
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.NotEqual(t, resp.Headers[auth.AuthorizationHeader], "")
	userResp := rest.Result[nyumpb.UserRegistrationResponse]{}
	err = json.Unmarshal(resp.Body, &userResp)
	assert.NoError(t, err)
	require.True(t, len(userResp.Data) > 0)
	userID := userResp.Data[0].UserId

	// Test login with the created user
	resp, err = checkFunc(
		ctx,
		http.HandlerFunc(LoginUser(u)),
		check.WithBody(
			`{
				"email": "testuser@nyum.space",
				"password": "testpassword"
			}`,
		),
		check.WithURLPath("/portal/login"),
		check.WithURLPattern("/portal/login"),
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEqual(t, resp.Headers[auth.AuthorizationHeader], "")

	sessionToken := resp.Headers[auth.AuthorizationHeader]

	m := auth.NewMiddleware(&d)
	resp, err = checkFunc(
		ctx,
		http.HandlerFunc(GetUser(u)),
		check.WithMiddlewares(m.AuthorizeSession),
		check.WithURLPath("/api/portal"),
		check.WithURLPattern("/api/portal"),
		check.WithHeaders(
			check.Header(auth.AuthorizationHeader, sessionToken),
		),
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.JSONEq(t,
		fmt.Sprintf(`{
				"data":[
					{"user_id":"%s",
	 				 "username":"testuser2",
 	 				 "email":"testuser@nyum.space"}
						],
				"message":"OK"
								}`, userID),
		string(resp.Body))

	// Test update user
	update, err := json.Marshal(nyum.UserUpdateRequest{
		UserID: userID,
		UserUpdateRequest: nyumpb.UserUpdateRequest{
			Username: "updated-username",
			Email:    "updated@nyum.space",
		},
	})
	require.NoError(t, err)
	resp, err = checkFunc(
		ctx,
		http.HandlerFunc(UpdateUser(u)),
		check.WithMiddlewares(m.AuthorizeSession),
		check.WithURLPath("/api/portal"),
		check.WithURLPattern("/api/portal"),
		check.WithHeaders(
			check.Header(auth.AuthorizationHeader, sessionToken),
		),
		check.WithBody(string(update)),
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	got, err := u.GetUser(ctx, &nyum.UserRequest{
		UserRequest: nyumpb.UserRequest{
			UserId: userID,
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, got.Email, "updated@nyum.space")

	resp, err = checkFunc(
		ctx,
		http.HandlerFunc(DeleteUser(u)),
		check.WithMiddlewares(
			m.AuthorizeSession,
		),
		check.WithURLPath("/api/portal"),
		check.WithURLPattern("/api/portal"),
		check.WithHeaders(
			check.Header(auth.AuthorizationHeader, sessionToken),
		),
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	got, err = u.GetUser(ctx, &nyum.UserRequest{
		UserRequest: nyumpb.UserRequest{
			UserId: userID,
		},
	})
	assert.NoError(t, err)
	assert.Nil(t, got)
}
