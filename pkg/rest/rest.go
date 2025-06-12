package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Result is a generic response structure for REST endpoints
type Result[T any] struct {
	Data    []T      `json:"data,omitempty"`
	Message string   `json:"message,omitempty"`
	Errors  []string `json:"errors,omitempty"`
}

// Add to rest.go
func ValidationFailed(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity) // 422
	json.NewEncoder(w).Encode(Result[any]{
		Message: "Validation failed",
		Errors:  []string{err.Error()},
	})
}

func InternalError(w http.ResponseWriter, err error) {
	sendJSON(w, http.StatusInternalServerError, Result[any]{
		Message: err.Error(),
		Errors:  []string{err.Error()},
	})
}

func BadRequest(w http.ResponseWriter, err error) {
	sendJSON(w, http.StatusBadRequest, Result[any]{
		Errors:  []string{err.Error()},
		Message: err.Error(),
	})
}

func Created[T any](w http.ResponseWriter, data ...T) {
	sendJSON(w, http.StatusCreated, Result[T]{
		Data:    append([]T(nil), data...),
		Message: "Created successfully",
	})
}

func OK[T any](w http.ResponseWriter, data ...T) {
	sendJSON(w, http.StatusOK, Result[T]{
		Data:    append([]T(nil), data...),
		Message: "OK",
	})
}

func Unauthorized(w http.ResponseWriter, err error) {
	sendJSON(w, http.StatusUnauthorized, Result[any]{
		Message: "Unauthorized",
		Errors:  []string{err.Error()},
	})
}

func sendJSON[T any](w http.ResponseWriter, status int, data T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ExtractPayload[T any](r *http.Request) (T, error) {
	var payload T
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return payload, fmt.Errorf("invalid JSON: %w", err)
	}

	return payload, nil
}
