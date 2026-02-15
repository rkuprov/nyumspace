package handlers

import (
	"fmt"
	"net/http"

	"go.temporal.io/sdk/client"

	"github.com/rkuprov/nyumspace/applications/nyumspace/business/workflows"
)

func Home() http.HandlerFunc {
	c, err := client.Dial(client.Options{
		Namespace: "nyumspace",
		HostPort:  "localhost:7233",
	})
	if err != nil {
		panic(err)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		opts := client.StartWorkflowOptions{
			ID:        "hello-workflow",
			TaskQueue: "api",
		}
		we, err := c.ExecuteWorkflow(r.Context(), opts, workflows.Hello, "roman")
		if err != nil {
			http.Error(w, "Failed to start workflow", http.StatusInternalServerError)
			return
		}
		var out string
		we.Get(r.Context(), &out)

		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "%s", out)
	}
}

func HealthFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func Hello() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(writer, "hello world!")
	}
}
