package workflows

import (
	"go.temporal.io/sdk/workflow"
)

func Hello(ctx workflow.Context, name string) (string, error) {
	return "Hello, " + name + "!", nil
}
