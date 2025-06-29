package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	errstk "github.com/nnishant776/errstack"
	"golang.org/x/sys/unix"
)

func init() {
	chainFmtOpts := errstk.DefaultChainErrorFormatter.Options()
	chainFmtOpts.ErrorSeparator = "; "
	chainFmtter := errstk.DefaultChainErrorFormatter.Copy()
	chainFmtter.SetOptions(chainFmtOpts)
	errstk.DefaultChainErrorFormatter = chainFmtter
}

func main() {
	ctx, cancelFn := signal.NotifyContext(context.Background(), []os.Signal{unix.SIGTERM, unix.SIGINT}...)
	defer cancelFn()

	cmd := newCLICmds()
	err := cmd.ExecuteContext(ctx)
	if err != nil && cmd.Flags().Lookup("verbose").Value.String() == "true" {
		fmt.Printf("Failed to execute command: %#4v\n", err)
	}
}
