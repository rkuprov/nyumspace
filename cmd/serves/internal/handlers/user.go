package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/rkuprov/nyumspace/cmd/serves/internal/users"
	"github.com/rkuprov/nyumspace/pkg/auth"
	"github.com/rkuprov/nyumspace/pkg/gen/nyumpb"
	"github.com/rkuprov/nyumspace/pkg/nyum"
	"github.com/rkuprov/nyumspace/pkg/rest"
)

// User Handlers

// RegisterUser creates a handler for user registration
func RegisterUser(u *users.Users) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := rest.ExtractPayload[nyumpb.UserRegistrationRequest](r)
		if err != nil {
			rest.ErrBadRequest(w, err)
			return
		}

		// Validate required fields
		if req.Username == "" || req.Email == "" || req.Password == "" {
			rest.ErrValidation(w, errors.New("must have username, email, and password"))
			return
		}

		resp, err := u.RegisterUser(r.Context(), &nyum.UserRegistrationRequest{
			UserRegistrationRequest: nyumpb.UserRegistrationRequest{
				Username: req.Username,
				Password: req.Password,
				Email:    req.Email,
			},
		})
		switch {
		case err == nil:
		case resp == nil:
			rest.ErrInternal(w, fmt.Errorf("failed to register user: %w", err))
			return
		default:
			rest.ErrBadRequest(w, err)
			return
		}

		loginReq := &nyumpb.UserLoginRequest{
			Email:    req.Email,
			Password: req.Password,
		}
		_, err = beginSession(r.Context(), u, loginReq, w)
		if err != nil {
			rest.ErrInternal(w, fmt.Errorf("failed to login after registration: %w", err))
			return
		}

		rest.Created(w, resp)
	}
}

// GetUser creates a handler for retrieving a user by ID
func GetUser(u *users.Users) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get(auth.UserIDHeader)
		resp, err := u.GetUser(r.Context(), &nyum.UserRequest{
			UserRequest: nyumpb.UserRequest{
				UserId: userID,
			},
		})
		if err != nil {
			if errors.Is(err, users.ErrNotFound) {
				// If user not found, return 404
				rest.ErrNotFound(w, err)
				return
			}
			rest.ErrInternal(w, fmt.Errorf("failed to get user: %w", err))
			return
		}

		rest.OK(w, resp)
	}
}

// UpdateUser creates a handler for updating a user
func UpdateUser(u *users.Users) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get(auth.UserIDHeader)
		req, err := rest.ExtractPayload[nyumpb.UserUpdateRequest](r)
		if err != nil {
			rest.ErrBadRequest(w, err)
			return
		}

		resp, err := u.UpdateUser(r.Context(), &nyum.UserUpdateRequest{
			UserID: userID,
			UserUpdateRequest: nyumpb.UserUpdateRequest{
				Username: req.Username,
				Email:    req.Email,
			},
		})
		if err != nil {
			rest.ErrInternal(w, fmt.Errorf("failed to update user: %w", err))
			return
		}

		rest.OK(w, resp)
	}
}

// DeleteUser creates a handler for deleting a user
func DeleteUser(u *users.Users) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get(auth.UserIDHeader)

		resp, err := u.DeleteUser(r.Context(), &nyum.UserDeleteRequest{
			UserDeleteRequest: nyumpb.UserDeleteRequest{
				UserId: userID,
			},
		})
		if err != nil {
			rest.ErrInternal(w, fmt.Errorf("failed to delete user: %w", err))
			return
		}

		_, err = u.LogoutUser(r.Context(), &nyum.UserLogoutRequest{
			UserLogoutRequest: nyumpb.UserLogoutRequest{
				SessionToken: r.Header.Get(auth.AuthorizationHeader),
			},
		})
		if err != nil {
			rest.ErrInternal(w, fmt.Errorf("failed to logout user after deletion: %w", err))
			return
		}

		rest.OK(w, resp)
	}
}

// LoginUser creates a handler for user login
func LoginUser(u *users.Users) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := rest.ExtractPayload[nyumpb.UserLoginRequest](r)
		if err != nil {
			rest.ErrBadRequest(w, err)
			return
		}

		// Validate required fields
		if req.Email == "" || req.Password == "" {
			rest.ErrValidation(w, errors.New("email and password are required"))
			return
		}

		resp, err := beginSession(r.Context(), u, &req, w)
		if err != nil {
			rest.ErrInternal(w, fmt.Errorf("failed to login: %w", err))
			return
		}

		rest.OK(w, &resp)
	}
}

// LogoutUser creates a handler for user logout
func LogoutUser(u *users.Users) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get session token from Authorization header
		sessionToken := r.Header.Get("Authorization")
		if sessionToken == "" {
			rest.ErrUnauthorized(w, errors.New("missing Authorization header"))
			return
		}

		// Remove "Bearer " prefix if present
		if len(sessionToken) > 7 && sessionToken[:7] == "Bearer " {
			sessionToken = sessionToken[7:]
		}

		resp, err := u.LogoutUser(r.Context(), &nyum.UserLogoutRequest{
			UserLogoutRequest: nyumpb.UserLogoutRequest{
				SessionToken: sessionToken,
			},
		})
		if err != nil {
			rest.ErrInternal(w, fmt.Errorf("failed to logout: %w", err))
			return
		}

		rest.OK(w, resp)
	}
}

func beginSession(
	ctx context.Context,
	u *users.Users,
	req *nyumpb.UserLoginRequest,
	w http.ResponseWriter,
) (*nyum.UserLoginResponse, error) {
	resp, err := u.LoginUser(ctx, &nyum.UserLoginRequest{
		UserLoginRequest: nyumpb.UserLoginRequest{
			Email:    req.Email,
			Password: req.Password,
		},
	})
	if err != nil {
		return nil, err
	}

	w.Header().Add("Authorization", resp.SessionToken)

	return resp, nil
}
