package k3d

import (
	"os"
	"path/filepath"

	errstk "github.com/nnishant776/errstack"
	"github.com/nnishant776/local-cluster/internal/tools"
	"github.com/nnishant776/local-cluster/pkg/model/cluster/k3s"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
)

func NewGencfgCommand(_ *k3s.ClusterConfig) *cobra.Command {
	gencfgCmd := &cobra.Command{
		Use:   "gencfg",
		Short: "Generate the cluster configuration",
		Long:  "Generate the cluster configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Extract the config file path
			outputPath := ""
			if outPathFlag := cmd.Flag("output-path"); outPathFlag != nil {
				if outPath := outPathFlag.Value.String(); outPath != "" {
					outputPath = outPath
				} else {
					if clusterCfg := cmd.Flag("cluster-config"); clusterCfg != nil {
						outputPath = clusterCfg.Value.String()
					}
				}
			}

			if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
				return errstk.New(err, errstk.WithStack())
			}

			dst, err := os.Create(outputPath)
			if err != nil {
				return errstk.New(err, errstk.WithStack())
			}
			defer dst.Close()

			err = unix.Dup2(int(dst.Fd()), int(os.Stdout.Fd()))
			if err != nil {
				return errstk.New(err, errstk.WithStack())
			}

			// Generate cluster configuration before creating the cluster
			helmfileCmd := tools.NewHelmfileCommand(nil)
			cmdArgs := []string{
				"--environment", "k3s",
				"template",
				"-l", "name=cluster",
				"--disable-force-update",
				"--state-values-set", "installed=true",
			}
			if deployCfg := cmd.Flag("deploy-config"); deployCfg != nil {
				cmdArgs = append(cmdArgs, "--state-values-set", "deploy-config="+deployCfg.Value.String())
			}
			if v := cmd.Flag("verbose"); v != nil && v.Value.String() == "true" {
				cmdArgs = append(cmdArgs, "--debug")
			}
			helmfileCmd.SetArgs(cmdArgs)
			if err := helmfileCmd.ExecuteContext(cmd.Context()); err != nil {
				return err
			}

			return nil
		},
	}

	gencfgCmd.Flags().StringP(
		"output-path", "o",
		"cluster/config.yaml",
		"Output path for the generate configuration",
	)

	return gencfgCmd
}
