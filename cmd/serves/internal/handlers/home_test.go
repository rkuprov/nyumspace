package handlers

import (
	"context"
	"fmt"
	"log"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rkuprov/nyumspace/cmd/serves/internal/sql"
	"github.com/rkuprov/nyumspace/pkg/daemon"
	"github.com/rkuprov/nyumspace/pkg/gen/nyumpb"
	"github.com/rkuprov/nyumspace/pkg/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// createTestUser inserts a user directly into the DB for test purposes
// and returns the user's ID (UUID string).
func createTestUser(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	userID := uuid.New().String()
	// Generate unique username/email for each test user to avoid conflicts if run in parallel or reused DB
	username := "owner-" + uuid.NewString()[:8]
	email := username + "@example.com"
	password := "password123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err, "Failed to hash password for test user")

	// Using the user schema's SQL structure directly
	insertUserSQL := `INSERT INTO users (id, name, email, password_hash) VALUES ($1, $2, $3, $4) RETURNING id;`
	var returnedUserID string
	err = pool.QueryRow(context.Background(), insertUserSQL, userID, username, email, string(hashedPassword)).Scan(&returnedUserID)
	require.NoError(t, err, "Failed to insert test user")
	require.Equal(t, userID, returnedUserID, "Returned user ID does not match generated user ID")
	return userID
}

func TestServerHandler_AddHome(t *testing.T) {
	pool := tests.DBForTest(t)
	dbname := pool.Config().ConnConfig.Database
	defer func() {
		if err := tests.RemoveDBForTest(dbname); err != nil {
			log.Printf("WARN: failed to remove test database: %v", err)
		}
	}()
	defer pool.Close()

	ownerID := createTestUser(t, pool)

	svs := NewServerHandler(daemon.Daemon{DB: pool})
	homeName := "Test Home " + uuid.NewString()[:4]
	req := &connect.Request[nyumpb.HomeCreationRequest]{
		Msg: &nyumpb.HomeCreationRequest{
			OwnerId:         ownerID,
			Name:            homeName,
			StreetAddress_1: "123 Main St", // Matches field name in home.go handler
			StreetAddress_2: "Apt 4B",
			City:            "Testville",
			State:           "TS",
			ZipCode:         "12345",
			Country:         "Testland",
			Description:     "A lovely test home.",
			Tags:            []string{"test", "cozy"},
			ImageUrl:        "http://example.com/image.jpg",
		},
	}

	resp, err := svs.AddHome(context.Background(), req)
	require.NoError(t, err, "AddHome request failed")
	require.NotNil(t, resp, "Result from AddHome is nil")
	require.NotNil(t, resp.Msg, "Result message from AddHome is nil")
	require.True(t, resp.Msg.GetSuccess(), "AddHome was not successful")

	assert.NotEmpty(t, resp.Msg.GetHomeId(), "Home ID in response is empty")
	_, err = uuid.Parse(resp.Msg.GetHomeId())
	require.NoError(t, err, "Home ID in response is not a valid UUID")
	assert.Equal(t, fmt.Sprintf("Home '%s' created successfully", homeName), resp.Msg.GetMessage())

	// Verify in DB
	var dbName string
	err = pool.QueryRow(context.Background(), "SELECT name FROM homes WHERE id = $1", resp.Msg.GetHomeId()).Scan(&dbName)
	require.NoError(t, err, "Failed to query created home from DB")
	assert.Equal(t, homeName, dbName, "Home name in DB does not match")
}

func TestServerHandler_GetHome(t *testing.T) {
	ctx := t.Context()
	pool := tests.DBForTest(t)
	dbname := pool.Config().ConnConfig.Database
	defer func() {
		if err := tests.RemoveDBForTest(dbname); err != nil {
			log.Printf("WARN: failed to remove test database: %v", err)
		}
	}()
	defer pool.Close()

	ownerID := createTestUser(t, pool)
	svs := NewServerHandler(daemon.Daemon{DB: pool})

	homeID := uuid.New().String()
	homeName := "Gettable Home"
	street1 := "456 Oak Ave"
	street2 := "Suite 100"
	city := "GetCity"
	state := "GS"
	zip := "54321"
	country := "GetCountry"
	description := "A home to be fetched"
	tags := []string{"get", "fetch"}
	imageURL := "http://example.com/gettable.jpg"

	_, err := pool.Exec(context.Background(), sql.AddHomeSQL,
		homeID, ownerID, homeName, street1, street2, city, state, zip, country, description, tags, imageURL)
	require.NoError(t, err, "Failed to insert test home for GetHome test")

	_, err = svs.AddHome(ctx, &connect.Request[nyumpb.HomeCreationRequest]{
		Msg: &nyumpb.HomeCreationRequest{
			OwnerId:         ownerID,
			Name:            homeName,
			StreetAddress_1: street1,
			StreetAddress_2: street2,
			City:            city,
			State:           state,
			ZipCode:         zip,
			Country:         country,
			Description:     description,
			Tags:            tags,
			ImageUrl:        imageURL,
		},
	})

	require.NoError(t, err, "AddHome request failed")

	req := &connect.Request[nyumpb.HomeRequest]{
		Msg: &nyumpb.HomeRequest{HomeId: homeID},
	}

	resp2, err := svs.GetHome(context.Background(), req)
	require.NoError(t, err, "GetHome request failed")
	require.NotNil(t, resp2, "Result from GetHome is nil")
	require.NotNil(t, resp2.Msg, "Result message from GetHome is nil")

	assert.Equal(t, homeID, resp2.Msg.GetHomeId())
	assert.Equal(t, ownerID, resp2.Msg.GetOwnerId())
	assert.Equal(t, homeName, resp2.Msg.GetName())

}

func TestServerHandler_DeleteHome(t *testing.T) {
	pool := tests.DBForTest(t)
	dbname := pool.Config().ConnConfig.Database
	defer func() {
		if err := tests.RemoveDBForTest(dbname); err != nil {
			log.Printf("WARN: failed to remove test database: %v", err)
		}
	}()
	defer pool.Close()

	ownerID := createTestUser(t, pool)
	svs := NewServerHandler(daemon.Daemon{DB: pool})

	homeID := uuid.New().String()
	// Insert a home to be deleted
	_, err := pool.Exec(context.Background(), sql.AddHomeSQL,
		homeID, ownerID, "Deletable Home", "101 Delete Dr", "", "DeleteCity", "DS", "00000", "DeleteCountry", "To be deleted", []string{"delete"}, "delete.jpg")
	require.NoError(t, err, "Failed to insert test home for DeleteHome test")

	req := &connect.Request[nyumpb.HomeDeleteRequest]{
		Msg: &nyumpb.HomeDeleteRequest{HomeId: homeID},
	}

	resp, err := svs.DeleteHome(context.Background(), req)
	require.NoError(t, err, "DeleteHome request failed")
	require.NotNil(t, resp, "Result from DeleteHome is nil")
	require.NotNil(t, resp.Msg, "Result message from DeleteHome is nil")
	assert.True(t, resp.Msg.GetSuccess(), "DeleteHome was not successful")
	assert.Equal(t, fmt.Sprintf("Home with ID %s deleted successfully", homeID), resp.Msg.GetMessage())

	// Verify in DB
	var count int
	err = pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM homes WHERE id = $1", homeID).Scan(&count)
	require.NoError(t, err, "Failed to query DB after delete attempt")
	assert.Equal(t, 0, count, "Home was not deleted from the database")
}
