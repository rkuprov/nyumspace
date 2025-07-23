package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ydb-platform/ydb-go-sdk/v3/log"
)

// Result is a generic response structure for REST endpoints
type Result[T any] struct {
	Data    []T      `json:"data,omitempty"`
	Message string   `json:"message,omitempty"`
	Errors  []string `json:"errors,omitempty"`
}

func (r *Result[T]) Unpack(input string) error {
	err := json.NewDecoder(strings.NewReader(input)).Decode(r)
	if err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	return nil
}

// Add to rest.go

func ErrNotFound(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	if encodeErr := json.NewEncoder(w).Encode(Result[any]{
		Message: "Not Found",
		Errors:  []string{err.Error()},
	}); encodeErr != nil {
		http.Error(w, encodeErr.Error(), http.StatusInternalServerError)
	}
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
	if encodeErr := json.NewEncoder(w).Encode(Result[any]{
		Message: "Validation failed",
		Errors:  []string{err.Error()},
	}); encodeErr != nil {
		http.Error(w, encodeErr.Error(), http.StatusInternalServerError)
	}
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
	if err := json.NewEncoder(w).Encode(Result[any]{
		Message: msg,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func OK[T any](w http.ResponseWriter, data ...T) {
	sendJSON(w, http.StatusOK, Result[T]{
		Data:    append([]T(nil), data...),
		Message: "OK",
	})
}

func ResultOK(w http.ResponseWriter, result Result[any]) {
	sendJSON(w, http.StatusOK, result)
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
	defer func() {
		err := r.Body.Close()
		if err != nil {
			log.Error(err)
		}
	}()
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return payload, fmt.Errorf("invalid JSON: %w", err)
	}

	return payload, nil
}
