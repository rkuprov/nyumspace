package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rkuprov/checkpoint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rkuprov/nyumspace/cmd/serves/internal/homes"
	"github.com/rkuprov/nyumspace/cmd/serves/internal/users"
	"github.com/rkuprov/nyumspace/pkg/auth"
	"github.com/rkuprov/nyumspace/pkg/config"
	"github.com/rkuprov/nyumspace/pkg/daemon"
	"github.com/rkuprov/nyumspace/pkg/gen/nyumpb"
	"github.com/rkuprov/nyumspace/pkg/nyum"
	"github.com/rkuprov/nyumspace/pkg/rest"
	"github.com/rkuprov/nyumspace/pkg/storage"
	"github.com/rkuprov/nyumspace/pkg/tests"
)

func TestImageHandlers(t *testing.T) {
	// Starting Initial setup for the test
	user := struct {
		Email    string
		Name     string
		Password string
	}{
		Email:    "testuseremail@nyum.space",
		Password: "testpassword",
	}
	cfg, err := config.NewConfig()
	require.NoError(t, err)
	db := tests.DBForTest(t)
	defer tests.CleanupTestDB(t, db)
	s, err := storage.NewStorageClient(t.Context(), cfg.S3Aws)
	require.NoError(t, err)
	h := &homes.Homes{
		DB:    db,
		Store: s,
	}
	m := auth.NewMiddleware(&daemon.Daemon{
		DB:      db,
		Storage: s,
	})
	u := &users.Users{db}
	u.RegisterUser(t.Context(), &nyum.UserRegistrationRequest{
		UserRegistrationRequest: nyumpb.UserRegistrationRequest{
			Username: user.Name,
			Email:    user.Email,
			Password: user.Password,
		},
	})
	loginResp, err := u.LoginUser(t.Context(), &nyum.UserLoginRequest{
		UserLoginRequest: nyumpb.UserLoginRequest{
			Email:    user.Email,
			Password: user.Password,
		},
	})
	require.NoError(t, err)
	homeResp, err := h.CreateHome(t.Context(), &nyum.HomeCreationRequest{
		UserID: loginResp.UserId,
		HomeCreationRequest: nyumpb.HomeCreationRequest{
			Name:            "Test Home",
			StreetAddress_1: "123 Test St",
			StreetAddress_2: "Apt 4B",
			City:            "Test City",
			State:           "TS",
			ZipCode:         "12345",
			Country:         "Testland",
			Description:     "A test home for image upload",
		},
	})
	homeID := homeResp.HomeId

	// Setup file attachment
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	payload := []byte("This is a test image file content")
	part, err := writer.CreateFormFile("image", "test-image.jpeg")
	require.NoError(t, err)
	_, err = io.Copy(part, bytes.NewReader(payload))
	require.NoError(t, err)
	err = writer.Close()

	// Test image upload
	endTest := checkpoint.Init(chi.NewRouter())
	endTest.RouteFunc = UploadHomeImage(h)
	endTest.WithMiddlewares(
		m.Session,
		m.AllowUser,
	)
	endTest.WithHeaders(
		checkpoint.Header(auth.AuthorizationHeader, loginResp.SessionToken),
		checkpoint.Header(auth.UserIDHeader, loginResp.UserId),
		checkpoint.Header("Content-Type", writer.FormDataContentType()),
	)
	endTest.URLPattern = "/homes/{home-id}/images"
	endTest.Path = fmt.Sprintf("/homes/%s/images", homeID)
	endTest.Method = http.MethodPost
	endTest.Body = io.NopCloser(&buf)
	r, err := endTest.Run(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, r.StatusCode)
	obj, err := h.Store.ListObjects(t.Context(), "images", loginResp.UserId)
	assert.NoError(t, err)
	assert.Len(t, obj, 1)
	assert.True(t, strings.HasSuffix(*obj[0].Key, "_test-image.jpeg"))

	var imgResp rest.Result[nyum.ImageCreateResponse]
	err = json.Unmarshal([]byte(r.Body.String()), &imgResp)
	require.NoError(t, err)
	require.NotNil(t, imgResp)

	// Test image retrieval
	imageID := imgResp.Data[0].ImageID

	retrieveTest := checkpoint.TestConfig{
		Router:     chi.NewRouter(),
		RouteFunc:  GetHomeImage(h),
		URLPattern: "/homes/{home-id}/images/{image-id}",
		Path:       fmt.Sprintf("/homes/%s/images/%s", homeID, imageID),
		Headers: map[string]string{
			auth.AuthorizationHeader: loginResp.SessionToken,
			auth.UserIDHeader:        loginResp.UserId},
		Middlewares: []func(http.Handler) http.Handler{
			m.Session,
			m.AllowUser,
		},
		Method: http.MethodGet,
	}
	out, err := retrieveTest.Run(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, out.StatusCode)
	assert.Equal(t, string(payload), out.Body.String())

	// Test image deletion
	deleteTest := checkpoint.TestConfig{
		Router:     chi.NewRouter(),
		RouteFunc:  DeleteHomeImage(h),
		URLPattern: "/homes/{home-id}/images/{image-id}",
		Path:       fmt.Sprintf("/homes/%s/images/%s", homeID, imageID),
		Headers: map[string]string{
			auth.AuthorizationHeader: loginResp.SessionToken,
			auth.UserIDHeader:        loginResp.UserId},
		Middlewares: []func(http.Handler) http.Handler{
			m.Session,
			m.AllowUser,
		},
		Method: http.MethodGet,
	}
	deleteOut, err := deleteTest.Run(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, out.StatusCode)
	res := rest.Result[any]{}
	err = res.Unpack(deleteOut.Body.String())
	assert.NoError(t, err)
	assert.Equal(t, "Image deleted successfully", res.Message)
}
