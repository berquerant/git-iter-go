package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/berquerant/git-iter-go/internal/cmd"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	return cmd.NewRootCmd().ExecuteContext(ctx)
}
