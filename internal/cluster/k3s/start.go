package k3d

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	errstk "github.com/nnishant776/errstack"
	"github.com/nnishant776/local-cluster/internal/utils"
	"github.com/nnishant776/local-cluster/pkg/model/cluster/k3s"
	"github.com/spf13/cobra"
)

func NewStartCommand(_ *k3s.ClusterConfig) *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the cluster",
		Long:  "Start the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Extract the config file path
			configPath := filepath.Join(utils.GetAppConfigDir(), "cluster", "config.yaml")

			newArgs := []string{"server"}
			if cmd.Flag("agent").Value.String() == "true" {
				newArgs = []string{"agent"}
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

			// Start the command
			if err := proc.Start(); err != nil {
				return errstk.New(fmt.Errorf("failed to start k3s: %w", err), errstk.WithStack())
			}

			// Record the PID of the started process
			runtimeDir := utils.GetAppRuntimeDir()
			if err := os.MkdirAll(runtimeDir, 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(
				filepath.Join(runtimeDir, "pid"),
				[]byte(strconv.FormatInt(int64(proc.Process.Pid), 10)),
				0o644,
			); err != nil {
				return err
			}

			// Wait for the process to complete
			return errstk.New(proc.Wait(), errstk.WithStack())
		},
	}

	startCmd.Flags().Bool(
		"agent",
		false,
		"Specify if agent should be created instead of a server",
	)

	return startCmd
}
