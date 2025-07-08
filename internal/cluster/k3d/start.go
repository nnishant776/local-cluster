package k3d

import (
	k3dcluster "github.com/k3d-io/k3d/v5/cmd/cluster"
	k3dclient "github.com/k3d-io/k3d/v5/pkg/client"
	"github.com/k3d-io/k3d/v5/pkg/runtimes"
	"github.com/k3d-io/k3d/v5/pkg/types"
	errstk "github.com/nnishant776/errstack"
	"github.com/nnishant776/local-cluster/pkg/model/cluster/k3d"
	"github.com/spf13/cobra"
)

func k3dStartCommand(cfg *k3d.ClusterConfig) *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the cluster",
		Long:  "Start the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Extract the config file path
			configPath := ""
			if clusterCfg := cmd.Flag("cluster-config"); clusterCfg != nil {
				configPath = clusterCfg.Value.String()
			}

			if _, err := k3dclient.ClusterGet(
				cmd.Context(), runtimes.SelectedRuntime, &types.Cluster{Name: cfg.Name},
			); err == nil {
				if len(args) <= 0 {
					args = []string{cfg.Name}
				}
			} else {
				if len(args) <= 0 {
					args = []string{"--config", configPath}
				}

				k3dCmd := k3dcluster.NewCmdClusterCreate()
				k3dCmd.SetArgs(args)
				err := k3dCmd.ExecuteContext(cmd.Context())
				if err != nil {
					return errstk.New(err, errstk.WithStack())
				}

				args = []string{cfg.Name}
			}

			k3dCmd := k3dcluster.NewCmdClusterStart()
			k3dCmd.SetArgs(args)
			err := k3dCmd.ExecuteContext(cmd.Context())
			if err != nil {
				err = errstk.New(err, errstk.WithStack())
			}

			return err
		},
	}

	return startCmd
}
