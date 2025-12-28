package tools

import (
	"os"
	"os/exec"

	"github.com/nnishant776/errstack"
	"github.com/spf13/cobra"
)

func NewHelmCommand() *cobra.Command {
	return &cobra.Command{
		Use:                "helm",
		Long:               "Run helm commands on the cluster",
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			proc := exec.Command("helm", args...)
			proc.Stdout, proc.Stderr = os.Stdout, os.Stderr
			if err := proc.Run(); err != nil {
				return errstack.New(err, errstack.WithStack())
			}

			return nil
		},
	}
}
