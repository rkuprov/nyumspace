package routes

import (
	"context"
	"net/http"
)

func Hello(_ context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	}
}
