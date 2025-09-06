package apps

import (
	"path/filepath"

	"github.com/nnishant776/local-cluster/config"
	"github.com/nnishant776/local-cluster/internal/tools"
	"github.com/nnishant776/local-cluster/internal/utils"
	"github.com/spf13/cobra"

	errstk "github.com/nnishant776/errstack"
)

func NewInstallCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "install",
		Short: "Install applications in the cluster",
		Long:  "Install applications in the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Extract the config file path
			deployConfig := filepath.Join(utils.GetAppConfigDir(), "config.yaml")

			cfg, err := config.Parse(deployConfig)
			if err != nil {
				return errstk.NewChainString(
					"app: failed to install application", errstk.WithStack(),
				).Chain(err)
			}

			helmfileCmd := tools.NewHelmfileCommand()

			// Chart installation args
			cmdArgs := []string{
				"-f", filepath.Join(utils.GetAppConfigDir(), "helmfile.yaml.gotmpl"),
				"--environment", cfg.Deployment.Environment.String(),
				"sync",
				"--disable-force-update",
			}

			// Add name filter if provided
			if appName := cmd.Flag("name"); appName != nil && appName.Value.String() != "" {
				cmdArgs = append(cmdArgs, "-l", "name="+appName.Value.String())
			}

			// Enable debug logging if verbose flag is specified
			if v := cmd.Flag("verbose"); v != nil && v.Value.String() == "true" {
				cmdArgs = append(cmdArgs, "--debug")
			}

			helmfileCmd.SetArgs(cmdArgs)
			if err := helmfileCmd.ExecuteContext(cmd.Context()); err != nil {
				return errstk.New(err, errstk.WithStack())
			}

			return nil
		},
	}

	return rootCmd
}
