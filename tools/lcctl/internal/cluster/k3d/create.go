package k3d

import (
	k3dcluster "github.com/k3d-io/k3d/v5/cmd/cluster"
	errstk "github.com/nnishant776/errstack"
	"github.com/nnishant776/local-cluster/pkg/model/cluster/k3d"
	"github.com/spf13/cobra"
)

func k3dCreateCommand(_ *k3d.ClusterConfig) *cobra.Command {
	k3dCmd := k3dcluster.NewCmdClusterCreate()
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create the cluster",
		Long:  "Create the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Extract the config file path
			configPath := ""
			if clusterCfg := cmd.Flag("cluster-config"); clusterCfg != nil {
				configPath = clusterCfg.Value.String()
			}

			if configPath != "" {
				newArgs := []string{"--config", configPath}
				for i := 0; i < len(args); i++ {
					arg := args[i]
					switch arg {
					case "-c", "--config":
						i += 2
					default:
						newArgs = append(newArgs, arg)
					}
				}
				args = newArgs
			}

			k3dCmd.SetArgs(args)
			err := k3dCmd.ExecuteContext(cmd.Context())
			if err != nil {
				err = errstk.New(err, errstk.WithStack())
			}

			return err
		},
	}

	return createCmd
}
