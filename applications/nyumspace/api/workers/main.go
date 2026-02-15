package main

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/rkuprov/nyumspace/applications/nyumspace/business/workflows"
)

func main() {
	client, err := client.Dial(client.Options{
		Namespace: "nyumspace",
		HostPort:  "localhost:7233",
	})
	if err != nil {
		panic(err)
	}

	w := worker.New(client, "api", worker.Options{})

	w.RegisterWorkflow(workflows.Hello)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		panic(err)
	}
}
