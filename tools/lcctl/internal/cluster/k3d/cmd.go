package k3d

import (
	"github.com/nnishant776/local-cluster/pkg/model/cluster/k3d"
	"github.com/spf13/cobra"
)

func NewK3DClusterCommand(cfg *k3d.ClusterConfig) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "cluster",
		Short: "Commands for cluster operation",
		Long:  "Commands for cluster operation",
	}

	rootCmd.AddCommand(k3dCreateCommand(cfg))
	rootCmd.AddCommand(k3dStartCommand(cfg))
	rootCmd.AddCommand(k3dStopCommand(cfg))
	rootCmd.AddCommand(k3dDestroyCommand(cfg))
	rootCmd.AddCommand(k3dGencfgCommand(cfg))
	rootCmd.AddCommand(k3dPurgeCommand(cfg))

	return rootCmd
}
