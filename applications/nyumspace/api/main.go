package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()
	err := SetupRoutes(r)
	if err != nil {
		panic(err)
	}

	err = http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}
}
