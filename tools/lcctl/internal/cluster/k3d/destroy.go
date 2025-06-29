package k3d

import (
	k3dcluster "github.com/k3d-io/k3d/v5/cmd/cluster"
	errstk "github.com/nnishant776/errstack"
	"github.com/nnishant776/local-cluster/pkg/model/cluster/k3d"
	"github.com/spf13/cobra"
)

func k3dDestroyCommand(cfg *k3d.ClusterConfig) *cobra.Command {
	k3dCmd := k3dcluster.NewCmdClusterDelete()
	destroyCmd := &cobra.Command{
		Use:   "destroy",
		Short: "Destroy the cluster",
		Long:  "Destroy the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			k3dCmd.SetArgs([]string{cfg.Name})
			err := k3dCmd.ExecuteContext(cmd.Context())
			if err != nil {
				err = errstk.New(err, errstk.WithStack())
			}

			return err
		},
	}

	return destroyCmd
}
