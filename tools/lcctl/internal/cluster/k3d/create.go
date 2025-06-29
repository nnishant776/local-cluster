package k3d

import (
	// "errors"
	"os"
	"os/exec"

	errstk "github.com/nnishant776/errstack"
	"github.com/nnishant776/local-cluster/pkg/model/cluster/k3d"
	"github.com/spf13/cobra"
)

func k3dCreateCommand(_ *k3d.ClusterConfig) *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create the cluster",
		Long:  "Create the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Extract the config file path
			configPath := ""
			if clusterCfg := cmd.Flag("cluster-config"); clusterCfg != nil {
				configPath = clusterCfg.Value.String()
			} else {
				return errstk.NewString(
					"cluster configuration file not found",
					errstk.WithStack(),
				)
			}

			// Define the command
			proc := exec.CommandContext(cmd.Context(), "k3d", "cluster", "create", "--config", configPath)

			// Connect outputs to the current process's outputs
			proc.Stdout = os.Stdout
			proc.Stderr = os.Stderr

			// Run the command till completion
			if proc != nil {
				if err := proc.Run(); err != nil {
					return errstk.New(err, errstk.WithStack())
				}
			}

			return nil
		},
	}

	return createCmd
}
