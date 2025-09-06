package k3d

import (
	"os"
	"path/filepath"

	errstk "github.com/nnishant776/errstack"
	"github.com/nnishant776/local-cluster/internal/tools"
	"github.com/nnishant776/local-cluster/internal/utils"
	"github.com/nnishant776/local-cluster/pkg/model/cluster/k3d"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
)

func NewGencfgCommand(_ *k3d.ClusterConfig) *cobra.Command {
	gencfgCmd := &cobra.Command{
		Use:   "gencfg",
		Short: "Generate the cluster configuration",
		Long:  "Generate the cluster configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Extract the config file path
			outputPath := filepath.Join(utils.GetAppConfigDir(), "cluster", "config.yaml")

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

			// Generate k3d configuration before creating the cluster
			helmfileCmd := tools.NewHelmfileCommand()
			cmdArgs := []string{
				"-f", filepath.Join(utils.GetAppConfigDir(), "helmfile.yaml.gotmpl"),
				"--environment", "k3d",
				"template",
				"-l", "name=cluster",
				"--disable-force-update",
				"--state-values-set", "installed=true",
			}
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

	return gencfgCmd
}
