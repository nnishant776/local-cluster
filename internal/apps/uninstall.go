package apps

import (
	"github.com/nnishant776/local-cluster/config"
	"github.com/nnishant776/local-cluster/internal/tools"
	"github.com/spf13/cobra"

	errstk "github.com/nnishant776/errstack"
)

func NewUninstallCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall applications in the cluster",
		Long:  "Uninstall applications in the cluster",
		// PreRunE: func(cmd *cobra.Command, args []string) error {
		// 	panic("TODO")
		// },
		RunE: func(cmd *cobra.Command, args []string) error {
			// Extract the config file path
			deployConfig := ""
			if deployCfg := cmd.Flag("deploy-config"); deployCfg != nil {
				deployConfig = deployCfg.Value.String()
			}

			cfg, err := config.Parse(deployConfig)
			if err != nil {
				return errstk.NewChainString(
					"app: failed to install application", errstk.WithStack(),
				).Chain(err)
			}

			helmfileCmd := tools.NewHelmfileCommand(nil)

			// Chart installation args
			cmdArgs := []string{
				"--environment", cfg.Deployment.Environment.String(),
				"destroy",
				"--disable-force-update",
				"--state-values-set", "deploy-config=" + deployConfig,
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
