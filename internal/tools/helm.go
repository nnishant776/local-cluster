package tools

import (
	"os"

	"github.com/spf13/cobra"
	helmcmd "helm.sh/helm/v4/pkg/cmd"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func NewHelmCommand(envConfig map[string]any) *cobra.Command {
	return &cobra.Command{
		Use:                "helm",
		Long:               "Run helm commands on the cluster",
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := helmcmd.NewRootCmd(os.Stderr, args, helmcmd.SetupLogging)
			if err != nil {
				return err
			}
			if v := cmd.Flag("verbose"); v != nil && v.Value.String() == "true" {
				args = append(args, "--debug")
			}
			c.SetArgs(args)
			return c.ExecuteContext(cmd.Context())
		},
	}
}
