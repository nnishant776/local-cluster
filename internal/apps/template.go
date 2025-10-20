package apps

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/nnishant776/local-cluster/config"
	"github.com/nnishant776/local-cluster/internal/tools"
	"github.com/nnishant776/local-cluster/internal/utils"
	"github.com/spf13/cobra"

	errstk "github.com/nnishant776/errstack"
)

func NewTemplateCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "template",
		Short: "Parse and print the chart template for the applications in the cluster",
		Long:  "Parse and print the chart template for the applications in the cluster",
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
				"-f", filepath.Join(filepath.Dir(deployConfig), "helmfile.yaml.gotmpl"),
				"--environment", cfg.Deployment.Environment.String(),
				"template",
				"--disable-force-update",
			}
			if kcfg := cmd.Flag("kubeconfig").Value.String(); kcfg != "" {
				cmdArgs = append(cmdArgs, "--kubeconfig", kcfg)
			}

			// Add name filter if provided
			if appName := cmd.Flag("name"); appName != nil && appName.Value.String() != "" {
				cmdArgs = append(cmdArgs, "-l", "name="+appName.Value.String())
			}
			if grpName := cmd.Flag("group"); grpName != nil && grpName.Value.String() != "" {
				cmdArgs = append(cmdArgs, "-l", "group="+grpName.Value.String())
			}
			if rawFilter := cmd.Flag("raw-filter"); rawFilter != nil && rawFilter.Value.String() != "" {
				rawFilterStr := rawFilter.Value.String()
				parts := strings.Split(rawFilterStr, "=")
				if len(parts) < 2 {
					return errstk.NewChainString(
						"app: failed to install application", errstk.WithStack(),
					).Chain(
						fmt.Errorf(
							"invalid filter format: '%s' must have 2 components joined by '='",
							rawFilterStr,
						),
					)
				}
				parts[0] = strings.TrimSpace(parts[0])
				parts[1] = strings.TrimSpace(parts[1])
				cmdArgs = append(cmdArgs, "-l", strings.Join(parts, "="))
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
