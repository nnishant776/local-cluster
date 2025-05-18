package k3d

import (
	"os"
	"os/exec"

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
			// Define the command
			proc := exec.CommandContext(cmd.Context(), "k3d", "cluster", "start", cfg.Name)

			// Connect outputs to the current process's outputs
			proc.Stdout = os.Stdout
			proc.Stderr = os.Stderr

			// Run the command till completion
			if err := proc.Run(); err != nil {
				return errstk.New(err, errstk.WithStack())
			}

			return nil
		},
	}

	return startCmd
}
