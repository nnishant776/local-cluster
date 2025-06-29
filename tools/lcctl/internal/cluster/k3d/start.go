package k3d

import (
	k3dcluster "github.com/k3d-io/k3d/v5/cmd/cluster"
	errstk "github.com/nnishant776/errstack"
	"github.com/nnishant776/local-cluster/pkg/model/cluster/k3d"
	"github.com/spf13/cobra"
)

func k3dStartCommand(cfg *k3d.ClusterConfig) *cobra.Command {
	k3dCmd := k3dcluster.NewCmdClusterStart()
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the cluster",
		Long:  "Start the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) <= 0 {
				args = []string{cfg.Name}
			}

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
