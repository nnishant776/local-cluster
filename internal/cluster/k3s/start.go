package k3d

import (
	"os"
	"os/exec"

	errstk "github.com/nnishant776/errstack"
	"github.com/nnishant776/local-cluster/pkg/model/cluster/k3s"
	"github.com/spf13/cobra"
)

func k3sStartCommand(_ *k3s.ClusterConfig) *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the cluster",
		Long:  "Start the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Extract the config file path
			configPath := ""
			if clusterCfg := cmd.Flag("cluster-config"); clusterCfg != nil {
				configPath = clusterCfg.Value.String()
			}

			newArgs := []string{"server"}
			if agentFlag := cmd.Flag("agent"); agentFlag != nil {
				if agentFlag.Value.String() == "true" {
					newArgs = []string{"agent"}
				}
			}

			if len(args) <= 0 {
				newArgs = append(newArgs, "--config", configPath)
			} else {
				newArgs = append(newArgs, args...)
			}

			// Define the command
			proc := exec.CommandContext(cmd.Context(), "k3s", newArgs...)

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

	startCmd.Flags().Bool(
		"agent",
		false,
		"Specify if agent should be created instead of a server",
	)

	return startCmd
}
