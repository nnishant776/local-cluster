package k3d

import (
	"os"
	"os/exec"

	errstk "github.com/nnishant776/errstack"
	"github.com/nnishant776/local-cluster/pkg/model/cluster/k3s"
	"github.com/spf13/cobra"
)

func k3sStopCommand(cfg *k3s.ClusterConfig) *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the cluster",
		Long:  "Stop the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Define the command
			proc := exec.CommandContext(cmd.Context(), "k3d", "cluster", "stop", cfg.Name)

			// Connect outputs to the current process's outputs
			proc.Stdout = os.Stdout
			proc.Stderr = os.Stderr

			// Run the command till completion
			if err := proc.Run(); err != nil {
				return errstk.New(err, errstk.WithTraceback())
			}

			return nil
		},
	}

	return startCmd
}
