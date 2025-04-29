package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"golang.org/x/sys/unix"
)

func main() {
	ctx, cancelFn := signal.NotifyContext(context.Background(), []os.Signal{unix.SIGTERM, unix.SIGINT}...)
	defer cancelFn()

	cmd := newCLICmds()
	err := cmd.ExecuteContext(ctx)
	if err != nil {
		fmt.Printf("Failed to execute command: %s", err)
	}
}
