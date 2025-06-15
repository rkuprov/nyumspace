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

func ErrNotFound(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(Result[any]{
		Message: "Not Found",
		Errors:  []string{err.Error()},
	})
}

func ErrUnauthorized(w http.ResponseWriter, err error) {
	sendJSON(w, http.StatusUnauthorized, Result[any]{
		Message: "ErrUnauthorized",
		Errors:  []string{err.Error()},
	})
}

func ErrValidation(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity) // 422
	json.NewEncoder(w).Encode(Result[any]{
		Message: "Validation failed",
		Errors:  []string{err.Error()},
	})
}

func ErrInternal(w http.ResponseWriter, err error) {
	sendJSON(w, http.StatusInternalServerError, Result[any]{
		Errors: []string{err.Error()},
	})
}

func ErrBadRequest(w http.ResponseWriter, err error) {
	sendJSON(w, http.StatusBadRequest, Result[any]{
		Message: err.Error(),
	})
}

func Created[T any](w http.ResponseWriter, data ...T) {
	sendJSON(w, http.StatusCreated, Result[T]{
		Data:    append([]T(nil), data...),
		Message: "Created",
	})
}

func NotFound(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Result[any]{
		Message: msg,
	})
}

func OK[T any](w http.ResponseWriter, data ...T) {
	sendJSON(w, http.StatusOK, Result[T]{
		Data:    append([]T(nil), data...),
		Message: "OK",
	})
}

func Mixed[T any](w http.ResponseWriter, status int, result Result[T]) {
	sendJSON(w, status, result)
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
