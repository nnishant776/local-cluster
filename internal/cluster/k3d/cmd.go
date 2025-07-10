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

	rootCmd.AddCommand(NewStartCommand(cfg))
	rootCmd.AddCommand(NewStopCommand(cfg))
	rootCmd.AddCommand(NewDestroyCommand(cfg))
	rootCmd.AddCommand(NewGencfgCommand(cfg))

	return rootCmd
}
