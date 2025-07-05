package k3d

import (
	"github.com/nnishant776/local-cluster/pkg/model/cluster/k3d"
	"github.com/spf13/cobra"
)

func k3dPurgeCommand(_ *k3d.ClusterConfig) *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "purge",
		Short: "Purge the cluster configuration",
		Long:  "Purge the cluster configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return startCmd
}
