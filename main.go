package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"os/signal"

	"github.com/nnishant776/local-cluster/internal/tools"

	errstk "github.com/nnishant776/errstack"

	"golang.org/x/sys/unix"
)

//go:embed all:deployment all:.devcontainer
var bundle embed.FS

func init() {
	chainFmtOpts := errstk.DefaultChainErrorFormatter.Options()
	chainFmtOpts.ErrorSeparator = "; "
	chainFmtter := errstk.DefaultChainErrorFormatter.Copy()
	chainFmtter.SetOptions(chainFmtOpts)
	errstk.DefaultChainErrorFormatter = chainFmtter
}

//go:generate go run ./prebuild.go download --report-progress=true
func main() {
	ctx, cancelFn := signal.NotifyContext(context.Background(), []os.Signal{unix.SIGTERM, unix.SIGINT}...)
	defer cancelFn()

	cmd := newCLICmds()
	switch os.Getenv("TOOL_MODE") {
	case "helm":
		cmd = tools.NewHelmCommand()
	case "helmfile":
		cmd = tools.NewHelmfileCommand()
	case "k9s":
		cmd = tools.NewK9SCommand()
	case "kubectl":
		cmd = tools.NewKubectlCommand()
	}

	err := cmd.ExecuteContext(ctx)
	if flg := cmd.Flag("verbose"); err != nil && flg != nil && flg.Value.String() == "true" {
		fmt.Printf("Failed to execute command: %#4v\n", err)
	}
}
