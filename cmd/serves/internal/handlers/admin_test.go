//go:build integration

package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"net/http"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rkuprov/checkpoint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ydb-platform/ydb-go-sdk/v3/log"

	"github.com/rkuprov/nyumspace/cmd/serves/internal/admin"
	"github.com/rkuprov/nyumspace/pkg/auth"
	"github.com/rkuprov/nyumspace/pkg/daemon"
	"github.com/rkuprov/nyumspace/pkg/nyum"
	"github.com/rkuprov/nyumspace/pkg/rest"
	"github.com/rkuprov/nyumspace/pkg/tests"
)

func createAdminUser(ctx context.Context, a *admin.Admin) string {
	var ret string
	err := a.DB.QueryRow(ctx, `insert into admins (id, name, email ) values ($1, $2, $3) returning id`,
		uuid.NewString(), `test-admin`, "admin@nyum.space").Scan(&ret)
	if err != nil {
		log.Error(err)
		return ""
	}
	return ret
}

func setupAdminTest(t *testing.T, ctx context.Context) (*admin.Admin, string, string) {
	db := tests.DBForTest(t)
	t.Cleanup(func() { tests.CleanupTestDB(t, db); fmt.Println("Test DB Cleaned up!!!") })
	a := admin.NewAdmin(daemon.Daemon{DB: db})
	adminID := createAdminUser(ctx, &a)
	if adminID == "" {
		t.Fatal("Failed to create admin user")
	}
	var session string
	err := a.DB.QueryRow(ctx,
		`insert into sessions (user_id, session_token, expires_at) values ($1, $2, $3) returning session_token`,
		adminID,
		uuid.NewString(),
		time.Now().Add(24*time.Hour)).Scan(&session)
	if err != nil {
		fmt.Println(err)
		log.Error(err)
		t.Fatal("Failed to create session for admin user")
	}
	return &a, adminID, session
}

func createUsers(ctx context.Context, a *admin.Admin, count int) iter.Seq[string] {
	return func(yield func(string) bool) {
		for i := range count {
			userID := createUser(ctx, a, i)
			if userID == "" {
				log.Error(errors.New("Failed to create user"))
				continue
			}
			if !yield(userID) {
				break
			}
		}
	}
}

func createUser(ctx context.Context, a *admin.Admin, num int) string {
	var ret string
	err := a.DB.QueryRow(ctx, `insert into users (id, name, email, password_hash) values ($1, $2, $3, $4) returning id`,
		uuid.NewString(),
		fmt.Sprintf("test-user-%d", num),
		fmt.Sprintf("test_email_%d@google.com", num),
		"hashed_password").Scan(&ret)
	if err != nil {
		log.Error(err)
		return ""
	}
	return ret
}

func crateHomes(ctx context.Context, a *admin.Admin, userID string, count int) iter.Seq[string] {
	return func(yield func(string) bool) {
		for i := range count {
			var homeID string
			err := a.DB.QueryRow(ctx, `insert into homes (
				id, 
				owner_id, 
				name,
				street_address_1,
				street_address_2,
				city,
				state,
				zip_code,
				country
				) values (
				$1, $2, $3, $4, $5, $6, $7, $8, $9) returning id`,
				uuid.NewString(),
				userID,
				fmt.Sprintf("Home of %s %d", userID, i),
				fmt.Sprintf("Street %d", i),
				fmt.Sprintf("Apt %d", i),
				fmt.Sprintf("City %d", i),
				fmt.Sprintf("State %d", i),
				fmt.Sprintf("Zip %d", i),
				fmt.Sprintf("Country %d", i),
			).Scan(&homeID)
			if err != nil {
				log.Error(err)
				continue
			}
			if homeID == "" {
				log.Error(errors.New("Failed to create home"))
				continue
			}
			if !yield(homeID) {
				break
			}
		}
	}
}

func TestGetAllUsers(t *testing.T) {
	ctx := context.Background()
	a, adminID, session := setupAdminTest(t, ctx)
	m := auth.NewMiddleware(&daemon.Daemon{DB: a.DB})

	var users []string
	for u := range createUsers(ctx, a, 10) {
		users = append(users, u)
	}

	tc := checkpoint.Init(chi.NewRouter())

	tc.RouteFunc = GetAllUsers(a)
	tc.URLPattern = "/admin/users"
	tc.Path = "/admin/users"
	tc.WithMiddlewares(m.Session, m.AllowAdmin)
	tc.WithHeaders(
		checkpoint.Header(auth.UserIDHeader, adminID),
		checkpoint.Header(auth.AuthorizationHeader, session))
	resp, err := tc.Run(ctx)
	assert.NoError(t, err, "Failed to get all users")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status OK")
	var out rest.Result[nyum.UserResponse]
	err = json.Unmarshal(resp.Body, &out)
	require.NoError(t, err, "Failed to unmarshal response")
	for _, user := range out.Data {
		assert.Contains(t, users, user.UserId, "User ID should be in the list of created users")
		assert.NotEmpty(t, user.Username, "Username should not be empty")
		assert.NotEmpty(t, user.Email, "Email should not be empty")
	}
}

func TestGetAllHomes(t *testing.T) {
	ctx := context.Background()
	a, adminID, session := setupAdminTest(t, ctx)
	m := auth.NewMiddleware(&daemon.Daemon{DB: a.DB})

	// Create some users and homes
	var homes []string
	for u := range createUsers(ctx, a, 10) {
		for h := range crateHomes(ctx, a, u, 3) {
			if h == "" {
				log.Error(errors.New("Failed to create home"))
				continue
			}
			homes = append(homes, h)
		}
	}

	tc := checkpoint.Init(chi.NewRouter())

	tc.RouteFunc = GetAllHomes(a)
	tc.URLPattern = "/admin/homes"
	tc.Path = "/admin/homes"
	tc.WithMiddlewares(m.Session, m.AllowAdmin)
	tc.WithHeaders(
		checkpoint.Header(auth.UserIDHeader, adminID),
		checkpoint.Header(auth.AuthorizationHeader, session))
	resp, err := tc.Run(ctx)
	assert.NoError(t, err, "Failed to get all homes")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status OK")
	var out rest.Result[nyum.HomeResponse]
	err = json.Unmarshal(resp.Body, &out)
	require.NoError(t, err, "Failed to unmarshal response")
	assert.Equal(t, len(homes), len(out.Data), "Expected number of homes")
	for _, home := range out.Data {
		assert.Contains(t, homes, home.GetHomeId(), "User ID should be in the list of created users")
		assert.NotEmpty(t, home.HomeId, "Home ID should not be empty")
		assert.NotEmpty(t, home.Name, "Home name should not be empty")
	}
}

func TestAdminDeleteUser(t *testing.T) {
	ctx := context.Background()
	a, adminID, session := setupAdminTest(t, ctx)
	m := auth.NewMiddleware(&daemon.Daemon{DB: a.DB})

	// Create a test user
	userID := createUser(ctx, a, 1)
	require.NotEmpty(t, userID, "Failed to create test user")

	tc := checkpoint.Init(chi.NewRouter())

	tc.RouteFunc = AdminDeleteUser(a)
	tc.URLPattern = "/admin/users/{userID}"
	tc.Path = fmt.Sprintf("/admin/users/%s", userID)
	tc.WithMiddlewares(m.Session, m.AllowAdmin)
	tc.WithHeaders(
		checkpoint.Header(auth.UserIDHeader, adminID),
		checkpoint.Header(auth.AuthorizationHeader, session))

	resp, err := tc.Run(ctx)
	assert.NoError(t, err, "Failed to delete user")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status OK")

	// Verify user was deleted
	var count int
	err = a.DB.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE id = $1", userID).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count, "User should be deleted")
}

func TestAdminDeleteHome(t *testing.T) {
	ctx := context.Background()
	a, adminID, session := setupAdminTest(t, ctx)
	m := auth.NewMiddleware(&daemon.Daemon{DB: a.DB})

	// Create a test user and home
	userID := createUser(ctx, a, 1)
	require.NotEmpty(t, userID, "Failed to create test user")

	var homeID string
	for h := range crateHomes(ctx, a, userID, 1) {
		homeID = h
		break
	}
	require.NotEmpty(t, homeID, "Failed to create test home")

	tc := checkpoint.Init(chi.NewRouter())

	tc.RouteFunc = AdminDeleteHome(a)
	tc.URLPattern = "/admin/homes/{homeID}"
	tc.Path = fmt.Sprintf("/admin/homes/%s", homeID)
	tc.WithMiddlewares(m.Session, m.AllowAdmin)
	tc.WithHeaders(
		checkpoint.Header(auth.UserIDHeader, adminID),
		checkpoint.Header(auth.AuthorizationHeader, session))

	resp, err := tc.Run(ctx)
	assert.NoError(t, err, "Failed to delete home")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status OK")

	// Verify home was deleted
	var count int
	err = a.DB.QueryRow(ctx, "SELECT COUNT(*) FROM homes WHERE id = $1", homeID).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count, "Home should be deleted")
}

func TestAdminGetHome(t *testing.T) {
	ctx := context.Background()
	a, adminID, session := setupAdminTest(t, ctx)
	m := auth.NewMiddleware(&daemon.Daemon{DB: a.DB})

	// Create a test user and home
	userID := createUser(ctx, a, 1)
	require.NotEmpty(t, userID, "Failed to create test user")

	var homeID string
	for h := range crateHomes(ctx, a, userID, 1) {
		homeID = h
		break
	}
	require.NotEmpty(t, homeID, "Failed to create test home")

	tc := checkpoint.Init(chi.NewRouter())

	tc.RouteFunc = AdminGetHome(a)
	tc.URLPattern = "/admin/homes/{home-id}"
	tc.Path = fmt.Sprintf("/admin/homes/%s", homeID)
	tc.WithMiddlewares(m.Session, m.AllowAdmin)
	tc.WithHeaders(
		checkpoint.Header(auth.UserIDHeader, adminID),
		checkpoint.Header(auth.AuthorizationHeader, session))

	resp, err := tc.Run(ctx)
	assert.NoError(t, err, "Failed to get home")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status OK")

	var out rest.Result[nyum.HomeResponse]
	err = json.Unmarshal(resp.Body, &out)
	require.NoError(t, err, "Failed to unmarshal response")
	require.Len(t, out.Data, 1, "Expected one home in response")

	home := out.Data[0]
	assert.Equal(t, homeID, home.HomeId, "Home ID should match")
	assert.Equal(t, userID, home.OwnerId, "Owner ID should match")
	assert.NotEmpty(t, home.Name, "Home name should not be empty")
}

func TestAdminGetHomesForUser(t *testing.T) {
	ctx := context.Background()
	a, adminID, session := setupAdminTest(t, ctx)
	m := auth.NewMiddleware(&daemon.Daemon{DB: a.DB})

	// Create a test user with multiple homes
	userID := createUser(ctx, a, 1)
	require.NotEmpty(t, userID, "Failed to create test user")

	var homes []string
	for h := range crateHomes(ctx, a, userID, 3) {
		homes = append(homes, h)
	}
	require.Len(t, homes, 3, "Should have created 3 homes")

	tc := checkpoint.Init(chi.NewRouter())

	tc.RouteFunc = AdminGetHomesForUser(a)
	tc.URLPattern = "/admin/users/{user-id}/homes"
	tc.Path = fmt.Sprintf("/admin/users/%s/homes", userID)
	tc.WithMiddlewares(m.Session, m.AllowAdmin)
	tc.WithHeaders(
		checkpoint.Header(auth.UserIDHeader, adminID),
		checkpoint.Header(auth.AuthorizationHeader, session))

	resp, err := tc.Run(ctx)
	assert.NoError(t, err, "Failed to get homes for user")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status OK")

	var out rest.Result[nyum.HomeResponse]
	err = json.Unmarshal(resp.Body, &out)
	require.NoError(t, err, "Failed to unmarshal response")
	assert.Len(t, out.Data, len(homes), "Expected 3 homes in response")

	for _, home := range out.Data {
		assert.Contains(t, homes, home.HomeId, "Home ID should be in created homes list")
		assert.Equal(t, userID, home.OwnerId, "Owner ID should match")
		assert.NotEmpty(t, home.Name, "Home name should not be empty")
	}
}

func TestAdminGetUser(t *testing.T) {
	ctx := context.Background()
	a, adminID, session := setupAdminTest(t, ctx)
	m := auth.NewMiddleware(&daemon.Daemon{DB: a.DB})

	// Create a test user
	userID := createUser(ctx, a, 1)
	require.NotEmpty(t, userID, "Failed to create test user")

	tc := checkpoint.Init(chi.NewRouter())

	tc.RouteFunc = AdminGetUser(a)
	tc.URLPattern = "/admin/users/{userID}"
	tc.Path = fmt.Sprintf("/admin/users/%s", userID)
	tc.WithMiddlewares(m.Session, m.AllowAdmin)
	tc.WithHeaders(
		checkpoint.Header(auth.UserIDHeader, adminID),
		checkpoint.Header(auth.AuthorizationHeader, session))

	resp, err := tc.Run(ctx)
	assert.NoError(t, err, "Failed to get user")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status OK")

	var out rest.Result[nyum.UserResponse]
	err = json.Unmarshal(resp.Body, &out)
	require.NoError(t, err, "Failed to unmarshal response")
	require.Len(t, out.Data, 1, "Expected one user in response")

	user := out.Data[0]
	assert.Equal(t, userID, user.UserId, "User ID should match")
	assert.Equal(t, "test-user-1", user.Username, "Username should match")
	assert.Equal(t, "test_email_1@google.com", user.Email, "Email should match")
}

func TestAdminGetUserNotFound(t *testing.T) {
	ctx := context.Background()
	a, adminID, session := setupAdminTest(t, ctx)
	m := auth.NewMiddleware(&daemon.Daemon{DB: a.DB})

	// Use a non-existent user ID
	nonExistentUserID := uuid.NewString()

	tc := checkpoint.Init(chi.NewRouter())

	tc.RouteFunc = AdminGetUser(a)
	tc.URLPattern = "/admin/users/{userID}"
	tc.Path = fmt.Sprintf("/admin/users/%s", nonExistentUserID)
	tc.WithMiddlewares(m.Session, m.AllowAdmin)
	tc.WithHeaders(
		checkpoint.Header(auth.UserIDHeader, adminID),
		checkpoint.Header(auth.AuthorizationHeader, session))

	resp, err := tc.Run(ctx)
	assert.NoError(t, err, "Failed to get user")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status OK")
}

func TestAdminGetHomeNotFound(t *testing.T) {
	ctx := context.Background()
	a, adminID, session := setupAdminTest(t, ctx)
	m := auth.NewMiddleware(&daemon.Daemon{DB: a.DB})

	// Use a non-existent home ID
	nonExistentHomeID := uuid.NewString()

	tc := checkpoint.Init(chi.NewRouter())

	tc.RouteFunc = AdminGetHome(a)
	tc.URLPattern = "/admin/homes/{home-id}"
	tc.Path = fmt.Sprintf("/admin/homes/%s", nonExistentHomeID)
	tc.WithMiddlewares(m.Session, m.AllowAdmin)
	tc.WithHeaders(
		checkpoint.Header(auth.UserIDHeader, adminID),
		checkpoint.Header(auth.AuthorizationHeader, session))

	resp, err := tc.Run(ctx)
	assert.NoError(t, err, "Failed to get home")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status Not Found")
}
