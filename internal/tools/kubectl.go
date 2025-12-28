package tools

import (
	"os"
	"os/exec"

	"github.com/nnishant776/errstack"
	"github.com/spf13/cobra"
)

func NewKubectlCommand() *cobra.Command {
	return &cobra.Command{
		Use:                "kubectl",
		Long:               "Run kubectl commands on the cluster",
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			proc := exec.Command("k9s", args...)
			proc.Stdout, proc.Stderr = os.Stdout, os.Stderr
			if err := proc.Run(); err != nil {
				return errstack.New(err, errstack.WithStack())
			}

			return nil
		},
	}
}
