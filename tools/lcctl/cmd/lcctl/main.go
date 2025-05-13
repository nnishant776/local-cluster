package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	errstk "github.com/nnishant776/errstack"
	"golang.org/x/sys/unix"
)

func main() {
	ctx, cancelFn := signal.NotifyContext(context.Background(), []os.Signal{unix.SIGTERM, unix.SIGINT}...)
	defer cancelFn()

	cmd := newCLICmds()
	err := cmd.ExecuteContext(ctx)
	if err != nil {
		fmt.Printf("Failed to execute command: %s\n", err)
		if e, ok := err.(errstk.Backtracer); ok &&
			cmd.Flags().Lookup("verbose").Value.String() == "true" {
			fmt.Printf("%s\n", e.Backtrace())
		}
	}
}
