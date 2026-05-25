package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/berquerant/yasp/internal"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGPIPE,
	)
	defer stop()
	return internal.Main(ctx, os.Stdout, os.Stdin, os.Args[1:]...)
}
