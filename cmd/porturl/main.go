package main

import (
	"context"
	"fmt"
	"os"

	"github.com/example/docker-port-scanner/internal/app"
	"github.com/example/docker-port-scanner/internal/dockercli"
)

func main() {
	if err := app.Run(context.Background(), os.Args[1:], os.Stdout, os.Stderr, dockercli.New()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
