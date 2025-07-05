package k3d

import (
	"github.com/nnishant776/local-cluster/internal/tools"
	"github.com/nnishant776/local-cluster/pkg/model/cluster/k3s"
	"github.com/spf13/cobra"
)

func k3sGencfgCommand(_ *k3s.ClusterConfig) *cobra.Command {
	gencfgCmd := &cobra.Command{
		Use:   "gencfg",
		Short: "Generate the cluster configuration",
		Long:  "Generate the cluster configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Generate cluster configuration before creating the cluster
			helmfileCmd := tools.NewHelmfileCommand(nil)
			helmfileCmd.SetArgs([]string{
				"--environment", "k3s",
				"template",
				"-l", "name=cluster",
				"--disable-force-update",
				"--state-values-set", "installed=true",
			})

			if err := helmfileCmd.ExecuteContext(cmd.Context()); err != nil {
				return err
			}

			return nil
		},
	}

	return gencfgCmd
}
