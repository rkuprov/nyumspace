package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rkuprov/nyumspace/pkg/daemon"
)

const (
	UserIDHeader = "NYUM-User-ID"
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

		r.Header.Add("NYUM-User-ID", userID)

		next.ServeHTTP(w, r)
	})
}

func checkToken(ctx context.Context, db *pgxpool.Pool, token string) (string, bool, error) {
	var userID string
	var expiresAt time.Time
	err := db.QueryRow(
		ctx,
		`SELECT user_id, expires_at FROM user_sessions WHERE session_token = $1`,
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
