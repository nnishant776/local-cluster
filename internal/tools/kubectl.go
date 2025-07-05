package tools

import (
	"github.com/spf13/cobra"
	kctlcmd "k8s.io/kubectl/pkg/cmd"
)

func NewKubectlCommand(envConfig map[string]any) *cobra.Command {
	return &cobra.Command{
		Use:  "kubectl",
		Long: "Run kubectl commands on the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := kctlcmd.NewDefaultKubectlCommand()
			c.SetArgs(args)
			return c.ExecuteContext(cmd.Context())
		},
		DisableFlagParsing: true,
	}
}
