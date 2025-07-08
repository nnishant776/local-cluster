package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"os/signal"

	"github.com/nnishant776/local-cluster/config"
	k3dc "github.com/nnishant776/local-cluster/internal/cluster/k3d"
	k3sc "github.com/nnishant776/local-cluster/internal/cluster/k3s"
	"github.com/nnishant776/local-cluster/internal/tools"
	"github.com/nnishant776/local-cluster/pkg/model"
	"github.com/nnishant776/local-cluster/pkg/model/cluster/k3d"
	"github.com/nnishant776/local-cluster/pkg/model/cluster/k3s"

	errstk "github.com/nnishant776/errstack"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
)

func init() {
	chainFmtOpts := errstk.DefaultChainErrorFormatter.Options()
	chainFmtOpts.ErrorSeparator = "; "
	chainFmtter := errstk.DefaultChainErrorFormatter.Copy()
	chainFmtter.SetOptions(chainFmtOpts)
	errstk.DefaultChainErrorFormatter = chainFmtter
}

//go:embed main.go
var fs embed.FS

func newCLICmds() *cobra.Command {
	rootCmd := rootCmd()
	rootCmd.AddCommand(toolsCmd())
	rootCmd.AddCommand(clusterCmd())
	return rootCmd
}

func rootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:              "lcctl",
		Short:            "lcctl is a tool controlling the ocal cluster deployment",
		Long:             "lcctl is a tool controlling the local cluster deployment",
		TraverseChildren: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:   true,
			DisableNoDescFlag:   false,
			DisableDescriptions: false,
			HiddenDefaultCmd:    false,
		},
	}

	rootCmd.PersistentFlags().String(
		"deploy-config", "config.yaml", "Deployment configuration file path",
	)
	rootCmd.PersistentFlags().Bool(
		"verbose", false, "Print verbose errors",
	)

	return rootCmd
}

func toolsCmd() *cobra.Command {
	toolCmd := &cobra.Command{
		Use:              "tools",
		Short:            "A collection of tools to help with cluster management",
		Long:             "A collection of tools to help with cluster management",
		TraverseChildren: true,
	}

	helmCmd := tools.NewHelmCommand(nil)
	helmfileCmd := tools.NewHelmfileCommand(nil)
	kubectlCmd := tools.NewKubectlCommand(nil)
	k9sCmd := tools.NewK9SCommand(nil)

	toolCmd.AddCommand(helmCmd)
	toolCmd.AddCommand(helmfileCmd)
	toolCmd.AddCommand(kubectlCmd)
	toolCmd.AddCommand(k9sCmd)

	return toolCmd
}

func clusterCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:                "cluster",
		Short:              "Commands for cluster operation",
		Long:               "Commands for cluster operation",
		TraverseChildren:   true,
		DisableFlagParsing: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:   true,
			DisableNoDescFlag:   false,
			DisableDescriptions: false,
			HiddenDefaultCmd:    false,
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Extract the config file path
			configPath := ""
			if deployConfig := cmd.Flag("deploy-config"); deployConfig != nil {
				configPath = deployConfig.Value.String()
			} else {
				return errstk.NewString("cluster: deployment config not found", errstk.WithStack())
			}

			cfg, err := config.Parse(configPath)
			if err != nil {
				return errstk.NewChainString(
					"cluster: command failed", errstk.WithStack(),
				).Chain(err)
			}

			switch cfg.Deployment.Environment {
			case model.K3D:
				if k3dClusterCfg, ok := cfg.Deployment.ClusterConfig.(*k3d.ClusterConfig); ok {
					cmd.AddCommand(k3dc.NewK3DClusterCommand(k3dClusterCfg).Commands()...)
				} else {
					return errstk.NewString(
						"invalid configuration: expected a k3d configuration", errstk.WithStack(),
					)
				}

			case model.K3S:
				if k3sClusterCfg, ok := cfg.Deployment.ClusterConfig.(*k3s.ClusterConfig); ok {
					cmd.AddCommand(k3sc.NewK3SClusterCommand(k3sClusterCfg).Commands()...)
				} else {
					return errstk.NewString(
						"invalid configuration: expected a k3s configuration", errstk.WithStack(),
					)
				}
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.RunE, cmd.PreRunE = nil, nil
			return cmd.Execute()
		},
	}

	rootCmd.PersistentFlags().String(
		"cluster-config", "cluster/config.yaml", "Cluster configuration file path",
	)

	return rootCmd
}

func main() {
	ctx, cancelFn := signal.NotifyContext(context.Background(), []os.Signal{unix.SIGTERM, unix.SIGINT}...)
	defer cancelFn()

	cmd := newCLICmds()
	err := cmd.ExecuteContext(ctx)
	if flg := cmd.Flag("verbose"); err != nil && flg != nil && flg.Value.String() == "true" {
		fmt.Printf("Failed to execute command: %#4v\n", err)
	}
}
