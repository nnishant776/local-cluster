package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"os/signal"

	errstk "github.com/nnishant776/errstack"

	"golang.org/x/sys/unix"
)

//go:embed all:deployment all:assets
var bundle embed.FS

func init() {
	chainFmtOpts := errstk.DefaultChainErrorFormatter.Options()
	chainFmtOpts.ErrorSeparator = "; "
	chainFmtter := errstk.DefaultChainErrorFormatter.Copy()
	chainFmtter.SetOptions(chainFmtOpts)
	errstk.DefaultChainErrorFormatter = chainFmtter
}

//go:generate go run ./prebuild.go download --component helm --report-progress=true --path assets
//go:generate go run ./prebuild.go download --component k3s --report-progress=true --path assets --tag v1.34.0
//go:generate go run ./prebuild.go download --component k9s --report-progress=true --path assets
//go:generate go run ./prebuild.go download --component kubectl --report-progress=true --path assets
//go:generate go run ./prebuild.go download --component helmfile --report-progress=true --path assets --tag v1.1.3
func main() {
	ctx, cancelFn := signal.NotifyContext(context.Background(), []os.Signal{unix.SIGTERM, unix.SIGINT}...)
	defer cancelFn()

	cmd := newCLICmds()
	err := cmd.ExecuteContext(ctx)
	if flg := cmd.Flag("verbose"); err != nil && flg != nil && flg.Value.String() == "true" {
		fmt.Printf("Failed to execute command: %#4v\n", err)
	}
}
