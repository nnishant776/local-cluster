package k3d

import (
	// "errors"
	// "fmt"
	// "io"
	// "os"
	// "path/filepath"

	errstk "github.com/nnishant776/errstack"
	"github.com/nnishant776/local-cluster/internal/tools"
	"github.com/nnishant776/local-cluster/pkg/model/cluster/k3s"
	"github.com/spf13/cobra"
)

func k3sGencfgCommand(_ *k3s.ClusterConfig) *cobra.Command {
	gencfgCmd := &cobra.Command{
		Use:   "gencfg",
		Short: "Generate the cluster configuration",
		Long:  "Generate the cluster configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Extract the config file path
			// outputPath := ""
			// if clusterCfg := cmd.Flag("output-path"); clusterCfg != nil {
			// 	outputPath = clusterCfg.Value.String()
			// } else {
			// 	return errstk.New(
			// 		errors.New("cluster configuration file not found"),
			// 		errstk.WithTraceback(),
			// 	)
			// }

			// Generate k3d configuration before creating the cluster
			helmfileCmd := tools.NewHelmfileCommand(nil)
			helmfileCmd.SetArgs([]string{
				"--environment", "k3s",
				"template",
				"-l", "name=cluster",
				"--disable-force-update",
				"--state-values-set", "installed=true",
			})
			if err := helmfileCmd.ExecuteContext(cmd.Context()); err != nil {
				return errstk.New(err, errstk.WithStack())
			}

			// if outputPath != "cluster/config.yaml" {
			// 	fmt.Printf("Copying cluster configuartion to destination: '%s'", outputPath)
			// 	src, err := os.Open("cluster/config.yaml")
			// 	if err != nil {
			// 		return errstk.New(err, errstk.WithTraceback())
			// 	}
			//
			// 	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
			// 		return errstk.New(err, errstk.WithTraceback())
			// 	}
			//
			// 	dst, err := os.OpenFile(outputPath, os.O_RDWR, 0o644)
			// 	if err != nil {
			// 		return errstk.New(err, errstk.WithTraceback())
			// 	}
			//
			// 	if _, err := io.Copy(dst, src); err != nil {
			// 		return errstk.New(err, errstk.WithTraceback())
			// 	}
			// }

			return nil
		},
	}

	// gencfgFlags := gencfgCmd.Flags()
	// gencfgFlags.StringP("output-path", "o", "cluster/config.yaml", "--output-path <filename>")

	return gencfgCmd
}
