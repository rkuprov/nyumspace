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

	"github.com/rkuprov/nyumspace/cmd/serves/internal/homes"
	"github.com/rkuprov/nyumspace/cmd/serves/internal/users"
	"github.com/rkuprov/nyumspace/pkg/auth"
	"github.com/rkuprov/nyumspace/pkg/daemon"
	"github.com/rkuprov/nyumspace/pkg/gen/nyumpb"
	"github.com/rkuprov/nyumspace/pkg/nyum"
	"github.com/rkuprov/nyumspace/pkg/rest"
	"github.com/rkuprov/nyumspace/pkg/tests"
)

func TestHomeEndpoints(t *testing.T) {
	ctx := context.Background()
	db := tests.DBForTest(t)
	defer tests.CleanupTestDB(t, db)

	d := daemon.Daemon{DB: db}
	u := users.NewUsers(&d)
	h := homes.NewHomes(&d)
	m := auth.NewMiddleware(&d)

	// First create a test user and get session token
	testConfig := check.Init(chi.NewRouter())
	testConfig.RouteFunc = RegisterUser(u)
	testConfig.Path = "/register"
	testConfig.Body = `{
		"username": "testuser",
		"email": "testuser@nyum.space",
		"password": "testpassword"
	}`

	resp, err := testConfig.Run(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	sessionToken := resp.Headers[auth.AuthorizationHeader]
	require.NotEmpty(t, sessionToken)

	userResp := rest.Result[nyumpb.UserRegistrationResponse]{}
	err = json.Unmarshal(resp.Body, &userResp)
	require.NoError(t, err)
	require.True(t, len(userResp.Data) > 0)
	userID := userResp.Data[0].UserId

	// Test creating a home
	testConfig.WithMiddlewares(m.Session)
	testConfig.WithHeaders(check.Header(auth.AuthorizationHeader, sessionToken))
	testConfig.RouteFunc = CreateHome(h)
	testConfig.Path = "/homes"
	testConfig.Body = `{
		"name": "Test Home",
		"street_address_1": "123 Test St",
		"street_address_2": "Apt 1",
		"city": "Test City",
		"state": "TS",
		"zip_code": "12345",
		"country": "USA",
		"description": "A test home",
		"tags": ["test", "home"],
		"image_url": "https://example.com/image.jpg"
	}`

	resp, err = testConfig.Run(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	homeResp := rest.Result[nyumpb.HomeCreationResponse]{}
	err = json.Unmarshal(resp.Body, &homeResp)
	assert.NoError(t, err)
	require.True(t, len(homeResp.Data) > 0)
	homeID := homeResp.Data[0].HomeId

	// Test getting the created home
	testConfig.RouteFunc = GetHome(h)
	testConfig.Path = fmt.Sprintf("/homes/%s", homeID)
	testConfig.URLPattern = "/homes/{home-id}"
	testConfig.Body = ""
	testConfig.WithMiddlewares(m.Session)
	testConfig.WithHeaders(check.Header(auth.AuthorizationHeader, sessionToken))

	resp, err = testConfig.Run(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	getHomeResp := rest.Result[nyumpb.HomeResponse]{}
	err = json.Unmarshal(resp.Body, &getHomeResp)
	assert.NoError(t, err)
	assert.True(t, len(getHomeResp.Data) > 0)
	assert.Equal(t, "Test Home", getHomeResp.Data[0].Name)
	assert.Equal(t, userID, getHomeResp.Data[0].OwnerId)
	// Test updating the home
	testConfig.RouteFunc = UpdateHome(h)
	testConfig.Path = fmt.Sprintf("/homes/%s", homeID)
	testConfig.URLPattern = "/homes/{home-id}"
	testConfig.Body = `{
		"name": "Updated Test Home",
		"street_address_1": "456 Updated St",
		"city": "Updated City",
		"state": "US",
		"zip_code": "54321",
		"country": "USA",
		"description": "An updated test home",
		"tags": ["updated", "test", "home"]
	}`
	testConfig.WithMiddlewares(m.Session)
	testConfig.WithHeaders(check.Header(auth.AuthorizationHeader, sessionToken))

	resp, err = testConfig.Run(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	// Verify the home was updated
	got, err := h.GetHome(ctx, &nyum.HomeRequest{
		HomeRequest: nyumpb.HomeRequest{
			HomeId: homeID,
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, "Updated Test Home", got.Name)
	assert.Equal(t, "456 Updated St", got.StreetAddress_1)
	assert.Equal(t, "Updated City", got.City)

	// Test getting all homes for user
	testConfig.RouteFunc = GetAllHomesForUser(h)
	testConfig.Path = "/homes/all"
	testConfig.Body = ""
	testConfig.WithMiddlewares(m.Session)
	testConfig.WithHeaders(check.Header(auth.AuthorizationHeader, sessionToken))

	resp, err = testConfig.Run(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	allHomesResp := rest.Result[nyum.HomeResponse]{}
	err = json.Unmarshal(resp.Body, &allHomesResp)
	assert.NoError(t, err)
	require.True(t, len(allHomesResp.Data) > 0)
	assert.Equal(t, homeID, allHomesResp.Data[0].HomeId)

	// Test deleting the home
	testConfig.RouteFunc = DeleteHome(h)
	testConfig.Path = fmt.Sprintf("/homes/%s", homeID)
	testConfig.URLPattern = "/homes/{home-id}"
	testConfig.Body = ""
	testConfig.WithMiddlewares(m.Session)
	testConfig.WithHeaders(check.Header(auth.AuthorizationHeader, sessionToken))

	resp, err = testConfig.Run(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify the home was deleted
	got, err = h.GetHome(ctx, &nyum.HomeRequest{
		HomeRequest: nyumpb.HomeRequest{
			HomeId: homeID,
		},
	})
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestHomeEndpointsUnauthorized(t *testing.T) {
	ctx := context.Background()
	db := tests.DBForTest(t)
	defer tests.CleanupTestDB(t, db)

	d := daemon.Daemon{DB: db}
	u := users.NewUsers(&d)
	h := homes.NewHomes(&d)
	m := auth.NewMiddleware(&d)

	// Create two test users
	testConfig := check.Init(chi.NewRouter())
	testConfig.RouteFunc = RegisterUser(u)
	testConfig.Path = "/register"
	testConfig.Body = `{
		"username": "user1",
		"email": "user1@nyum.space",
		"password": "password1"
	}`

	resp, err := testConfig.Run(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	sessionToken1 := resp.Headers[auth.AuthorizationHeader]
	require.NotEmpty(t, sessionToken1)

	// Create second user
	testConfig.Body = `{
		"username": "user2",
		"email": "user2@nyum.space",
		"password": "password2"
	}`

	resp, err = testConfig.Run(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	sessionToken2 := resp.Headers[auth.AuthorizationHeader]
	require.NotEmpty(t, sessionToken2)

	// User 1 creates a home
	testConfig.WithMiddlewares(m.Session)
	testConfig.WithHeaders(check.Header(auth.AuthorizationHeader, sessionToken1))
	testConfig.RouteFunc = CreateHome(h)
	testConfig.Path = "/homes"
	testConfig.Body = `{
		"name": "User1 Home",
		"street_address_1": "123 User1 St",
		"city": "User1 City",
		"state": "U1",
		"zip_code": "11111",
		"country": "USA",
		"description": "User1's home"
	}`

	resp, err = testConfig.Run(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	homeResp := rest.Result[nyumpb.HomeCreationResponse]{}
	err = json.Unmarshal(resp.Body, &homeResp)
	require.NoError(t, err)
	require.True(t, len(homeResp.Data) > 0)
	homeID := homeResp.Data[0].HomeId

	// User 2 tries to access User 1's home (should be unauthorized)
	testConfig.RouteFunc = GetHome(h)
	testConfig.Path = fmt.Sprintf("/homes/%s", homeID)
	testConfig.URLPattern = "/homes/{home-id}"
	testConfig.Body = ""
	testConfig.WithMiddlewares(m.Session)
	testConfig.WithHeaders(check.Header(auth.AuthorizationHeader, sessionToken2))

	resp, err = testConfig.Run(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// User 2 tries to update User 1's home (should be unauthorized)
	testConfig.RouteFunc = UpdateHome(h)
	testConfig.Path = fmt.Sprintf("/homes/%s", homeID)
	testConfig.Body = `{
		"name": "Hacked Home",
		"description": "This should not work"
	}`
	testConfig.WithMiddlewares(m.Session)
	testConfig.WithHeaders(check.Header(auth.AuthorizationHeader, sessionToken2))

	resp, err = testConfig.Run(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// User 2 tries to delete User 1's home (should be unauthorized)
	testConfig.RouteFunc = DeleteHome(h)
	testConfig.Path = fmt.Sprintf("/homes/%s", homeID)
	testConfig.Body = ""
	testConfig.WithMiddlewares(m.Session)
	testConfig.WithHeaders(check.Header(auth.AuthorizationHeader, sessionToken2))

	resp, err = testConfig.Run(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
