package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rkuprov/nyumspace/scratch/pkg/api/rest"
	"github.com/rkuprov/nyumspace/scratch/pkg/app/daemon"
)

const (
	UserIDHeader        = "NYUM-User-ID"
	AuthorizationHeader = "Authorization"
)

type Middleware struct {
	d *daemon.Daemon
}

func NewMiddleware(d *daemon.Daemon) *Middleware {
	return &Middleware{
		d: d,
	}
}

func (m *Middleware) Session(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID, valid, err := checkToken(r.Context(), m.d.DB, token)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if !valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		r.Header.Add(UserIDHeader, userID)

		next.ServeHTTP(w, r)
	})
}
func (m *Middleware) AllowUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get(UserIDHeader)
		if userID == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		err := checkUser(r.Context(), m.d.DB, userID)
		switch {
		case err == nil:
		case err.Error() == "user not found":
			rest.ErrUnauthorized(w, err)
			return
		default:
			rest.ErrInternal(w, err)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// AllowAdmin middleware ensures that the user is an admin
func (m *Middleware) AllowAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get(UserIDHeader)
		if userID == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		err := checkAdmin(r.Context(), m.d.DB, userID)
		switch {
		case err == nil:
		case err.Error() == "user not found":
			rest.ErrUnauthorized(w, err)
			return
		default:
			rest.ErrInternal(w, err)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func checkAdmin(ctx context.Context, db *pgxpool.Pool, userID string) error {
	var isAdmin bool
	err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM admins WHERE id = $1)`, userID).Scan(&isAdmin)

	switch {
	case err == nil:
		return nil
	case errors.Is(err, pgx.ErrNoRows):
		return errors.New("user not found")
	default:
		return err
	}
}
func checkUser(ctx context.Context, db *pgxpool.Pool, userID string) error {
	var ok bool
	err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, userID).Scan(&ok)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("user not found")
		}
		return err
	}
	return nil
}

func checkToken(ctx context.Context, db *pgxpool.Pool, token string) (string, bool, error) {
	var userID string
	var expiresAt time.Time
	err := db.QueryRow(
		ctx,
		`SELECT user_id, expires_at FROM sessions WHERE session_token = $1`,
		token).Scan(&userID, &expiresAt)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return "", false, err // Token does not exist
		}
		return "", false, nil // Token does not exist, but no error
	}

	if expiresAt.Before(time.Now()) {
		return "", false, nil // Token exists but is expired
	}

	return userID, true, nil
}
