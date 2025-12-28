package tools

import (
	"os"
	"os/exec"

	"github.com/nnishant776/errstack"
	"github.com/spf13/cobra"
)

func NewHelmfileCommand() *cobra.Command {
	return &cobra.Command{
		Use:                "helmfile",
		Long:               "Run helmfile commands on the cluster",
		DisableFlagParsing: true,
		RunE: func(c *cobra.Command, args []string) error {
			proc := exec.Command("helmfile", args...)
			proc.Stdout, proc.Stderr = os.Stdout, os.Stderr
			if err := proc.Run(); err != nil {
				return errstack.New(err, errstack.WithStack())
			}

			return nil
		},
	}
}
