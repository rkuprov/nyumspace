//go:build integration

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	check "github.com/rkuprov/checkpoint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rkuprov/nyumspace/cmd/serves/internal/users"
	"github.com/rkuprov/nyumspace/pkg/auth"
	"github.com/rkuprov/nyumspace/pkg/daemon"
	"github.com/rkuprov/nyumspace/pkg/gen/nyumpb"
	"github.com/rkuprov/nyumspace/pkg/nyum"
	"github.com/rkuprov/nyumspace/pkg/rest"
	"github.com/rkuprov/nyumspace/pkg/tests"
)

func TestUserEndpoints(t *testing.T) {
	ctx := context.Background()
	db := tests.DBForTest(t)
	defer tests.CleanupTestDB(t, db)

	d := daemon.Daemon{DB: db}
	u := users.NewUsers(&d)

	testConfig := check.Init(chi.NewRouter())
	testConfig.RouteFunc = RegisterUser(u)
	testConfig.Path = "/register"
	testConfig.SetBodyString(`{
	"username": "testuser",
	"email": "testuser@nyum.space",
	"password": "testpassword"
	}`)
	// Create a test user
	resp, err := testConfig.Run(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.NotEqual(t, resp.Headers[auth.AuthorizationHeader], "")
	userResp := rest.Result[nyumpb.UserRegistrationResponse]{}
	err = json.Unmarshal(resp.Body, &userResp)
	assert.NoError(t, err)
	require.True(t, len(userResp.Data) > 0)
	userID := userResp.Data[0].UserId

	// Test login with the created user
	testConfig.Path = "/portal/login"
	testConfig.SetBodyString(`{
				"email": "testuser@nyum.space",
				"password": "testpassword"
			}`)
	testConfig.Path = "/portal/login"

	testConfig.RouteFunc = LoginUser(u)
	resp, err = testConfig.Run(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEqual(t, resp.Headers[auth.AuthorizationHeader], "")

	m := auth.NewMiddleware(&d)
	sessionToken := resp.Headers[auth.AuthorizationHeader]
	testConfig.WithMiddlewares(m.Session)
	testConfig.Path = "/portal/portal"
	testConfig.WithHeaders(check.Header(auth.AuthorizationHeader, sessionToken))

	testConfig.RouteFunc = GetUser(u)
	resp, err = testConfig.Run(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.JSONEq(t,
		fmt.Sprintf(`{
				"data":[
					{"user_id":"%s",
	 				 "username":"testuser",
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

	testConfig.RouteFunc = UpdateUser(u)
	testConfig.Path = "/portal/portal"
	testConfig.WithHeaders(check.Header(auth.AuthorizationHeader, sessionToken))
	testConfig.SetBodyString(string(update))
	resp, err = testConfig.Run(ctx)
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

	testConfig.Path = "/portal/portal"
	testConfig.RouteFunc = DeleteUser(u)
	testConfig.WithHeaders(check.Header(auth.AuthorizationHeader, sessionToken))
	testConfig.WithMiddlewares(m.Session)
	resp, err = testConfig.Run(ctx)

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
