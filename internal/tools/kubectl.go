package tools

import (
	"os"

	"github.com/spf13/cobra"
	kctlcmd "k8s.io/kubectl/pkg/cmd"
)

func NewKubectlCommand() *cobra.Command {
	return &cobra.Command{
		Use:                "kubectl",
		Long:               "Run kubectl commands on the cluster",
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			os.Args = append([]string{cmd.Name()}, args...)
			c := kctlcmd.NewDefaultKubectlCommand()
			return c.ExecuteContext(cmd.Context())
		},
	}
}
