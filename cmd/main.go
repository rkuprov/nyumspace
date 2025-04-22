package main

import (
	"context"
	"errors"
	"fmt"
	"nyum/pkg/daemon"
	"time"
)

func main() {
	daemon.Run(func(ctx context.Context) error {
		fmt.Println("my awesome work function")
		time.Sleep(2 * time.Second)
		return errors.New("something bad happened")
	})
}
