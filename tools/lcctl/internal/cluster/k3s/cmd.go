package k3d

import (
	"github.com/nnishant776/local-cluster/pkg/model/cluster/k3s"
	"github.com/spf13/cobra"
)

func NewK3SClusterCommand(cfg *k3s.ClusterConfig) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "cluster",
		Short: "Commands for cluster operation",
		Long:  "Commands for cluster operation",
	}

	rootCmd.AddCommand(k3sCreateCommand(cfg))
	rootCmd.AddCommand(k3sStartCommand(cfg))
	rootCmd.AddCommand(k3sStopCommand(cfg))
	rootCmd.AddCommand(k3sDestroyCommand(cfg))
	rootCmd.AddCommand(k3sGencfgCommand(cfg))
	rootCmd.AddCommand(k3sPurgeCommand(cfg))

	return rootCmd
}
