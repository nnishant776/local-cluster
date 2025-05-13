package k3d

import (
	"os"

	"github.com/nnishant776/local-cluster/pkg/model/cluster/k3s"
	"github.com/spf13/cobra"
)

func k3sPurgeCommand(_ *k3s.ClusterConfig) *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "purge",
		Short: "Purge the cluster configuration",
		Long:  "Purge the cluster configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return os.RemoveAll("/var/lib/rancher/k3s")
		},
	}

	return startCmd
}
