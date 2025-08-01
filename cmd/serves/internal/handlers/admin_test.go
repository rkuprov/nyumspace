//go:build unit

package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rkuprov/checkpoint"
	"github.com/stretchr/testify/assert"

	"github.com/rkuprov/nyumspace/cmd/serves/internal/admin"
	"github.com/rkuprov/nyumspace/pkg/auth"
)

func TestAdminServerValidationErrors(t *testing.T) {
	mockA := admin.NewErrorMock("internal mock error")

	// Call the handler with the mock admin server
	handler := GetAllUsers(mockA)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/admin/users", nil)
	handler(w, r)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, `{"errors":["failed to get users: internal mock error"]}`, strings.TrimSpace(w.Body.String()))

	r = httptest.NewRequest("GET", "/admin/homes", nil)
	handler = GetAllHomes(mockA)
	w = httptest.NewRecorder()
	handler(w, r)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, `{"errors":["failed to get homes: internal mock error"]}`, strings.TrimSpace(w.Body.String()))

	r = httptest.NewRequest("DELETE", "/admin/users/user123", nil)
	handler = AdminDeleteUser(mockA)
	w = httptest.NewRecorder()
	handler(w, r)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Equal(t, `{"message":"Validation failed","errors":["userID is required"]}`, strings.TrimSpace(w.Body.String()))

	r = httptest.NewRequest("DELETE", "/admin/homes/home123", nil)
	handler = AdminDeleteHome(mockA)
	w = httptest.NewRecorder()
	handler(w, r)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Equal(t, `{"message":"Validation failed","errors":["homeID is required"]}`, strings.TrimSpace(w.Body.String()))

	r = httptest.NewRequest("GET", "/admin/homes/home123", nil)
	handler = AdminGetHome(mockA)
	w = httptest.NewRecorder()
	handler(w, r)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Equal(t, `{"message":"Validation failed","errors":["homeID is required"]}`, strings.TrimSpace(w.Body.String()))

	r = httptest.NewRequest("GET", "/admin/users/user123/homes", nil)
	handler = AdminGetHomesForUser(mockA)
	w = httptest.NewRecorder()
	handler(w, r)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Equal(t, `{"message":"Validation failed","errors":["userID is required"]}`, strings.TrimSpace(w.Body.String()))

	r = httptest.NewRequest("GET", "/admin/users/user123", nil)
	handler = AdminGetUser(mockA)
	w = httptest.NewRecorder()
	handler(w, r)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Equal(t, `{"message":"Validation failed","errors":["userID is required"]}`, strings.TrimSpace(w.Body.String()))
}

func TestAdminServerInternalErrors(t *testing.T) {
	mockA := admin.NewErrorMock("internal mock error")

	tc := &checkpoint.TestConfig{
		Router:     chi.NewRouter(),
		RouteFunc:  AdminGetHome(mockA),
		URLPattern: "/admin/homes/{home-id}",
		Path:       "/admin/homes/home123",
		Headers: map[string]string{
			"Content-Type":    "application/json",
			auth.UserIDHeader: "user123",
		},
	}

	out, err := tc.Run(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, out.StatusCode)
	assert.Equal(t, `{"errors":["failed to get home: internal mock error"]}`, strings.TrimSpace(out.Body.String()))

	tc = &checkpoint.TestConfig{
		Router:     chi.NewRouter(),
		RouteFunc:  AdminGetUser(mockA),
		URLPattern: "/admin/{user-id}",
		Path:       "/admin/user123",
	}
	out, err = tc.Run(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, out.StatusCode)
	assert.Equal(t, `{"errors":["failed to get user: internal mock error"]}`, strings.TrimSpace(out.Body.String()))

	tc = &checkpoint.TestConfig{
		Router:     chi.NewRouter(),
		RouteFunc:  AdminGetHomesForUser(mockA),
		URLPattern: "/admin/{user-id}/homes",
		Path:       "/admin/user123/homes",
	}
	out, err = tc.Run(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, out.StatusCode)
	assert.Equal(t, `{"errors":["failed to get homes for user: internal mock error"]}`, strings.TrimSpace(out.Body.String()))

	tc = &checkpoint.TestConfig{
		Router:     chi.NewRouter(),
		RouteFunc:  AdminDeleteUser(mockA),
		URLPattern: "/admin/{user-id}",
		Path:       "/admin/user123",
	}
	out, err = tc.Run(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, out.StatusCode)
	assert.Equal(t, `{"errors":["failed to delete user: internal mock error"]}`, strings.TrimSpace(out.Body.String()))

	tc = &checkpoint.TestConfig{
		Router:     chi.NewRouter(),
		RouteFunc:  AdminDeleteHome(mockA),
		URLPattern: "/admin/homes/{home-id}",
		Path:       "/admin/homes/home123",
	}
	out, err = tc.Run(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, out.StatusCode)
	assert.Equal(t, `{"errors":["failed to delete home: internal mock error"]}`, strings.TrimSpace(out.Body.String()))
}
