package main

import (
	"errors"
	"github.com/nnishant776/local-cluster/internal/tools"
	"github.com/spf13/cobra"

	errstk "github.com/nnishant776/errstack"
	"github.com/nnishant776/local-cluster/config"
	k3dc "github.com/nnishant776/local-cluster/internal/cluster/k3d"
	k3sc "github.com/nnishant776/local-cluster/internal/cluster/k3s"
	"github.com/nnishant776/local-cluster/pkg/model"
	"github.com/nnishant776/local-cluster/pkg/model/cluster/k3d"
	"github.com/nnishant776/local-cluster/pkg/model/cluster/k3s"
)

func newCLICmds() *cobra.Command {
	rootCmd := rootCmd()
	rootCmd.AddCommand(toolsCmd())
	rootCmd.AddCommand(clusterCmd())
	return rootCmd
}

func rootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "lcctl",
		Short: "lcctl is a tool controlling the local cluster deployment",
		Long:  "lcctl is a tool controlling the local cluster deployment",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:   true,
			DisableNoDescFlag:   false,
			DisableDescriptions: false,
			HiddenDefaultCmd:    false,
		},
	}

	rootCmd.PersistentFlags().StringP(
		"deploy-config", "d", "config.yaml", "--deploy-config <filename>",
	)

	return rootCmd
}

func toolsCmd() *cobra.Command {
	toolCmd := &cobra.Command{
		Use:   "tools",
		Short: "A collection of tools to help with cluster management",
		Long:  "A collection of tools to help with cluster management",
	}

	helmCmd := tools.NewHelmCommand(nil)
	helmfileCmd := tools.NewHelmfileCommand(nil)
	shellCmd := tools.NewShellCommand(nil)
	kubectlCmd := tools.NewKubectlCommand(nil)

	toolCmd.AddCommand(helmCmd)
	toolCmd.AddCommand(helmfileCmd)
	toolCmd.AddCommand(shellCmd)
	toolCmd.AddCommand(kubectlCmd)

	return toolCmd
}

func clusterCmd() *cobra.Command {
	clusterCmdsAdded := false

	rootCmd := &cobra.Command{
		Use:   "cluster",
		Short: "Commands for cluster operation",
		Long:  "Commands for cluster operation",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			if clusterCmdsAdded {
				return nil
			}
			clusterCmdsAdded = true
			return addClusterCmds(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// cmd.Run = nil
			cmd.RunE = nil
			// cmd.PreRun = nil
			cmd.PreRunE = nil

			if len(args) == 0 {
				return cmd.Help()
			}

			return cmd.Execute()
		},
	}

	rootCmd.SetHelpFunc(func(c *cobra.Command, s []string) {
		if !clusterCmdsAdded {
			clusterCmdsAdded = true
			addClusterCmds(c)
		}

		c.SetHelpFunc(nil)
		c.RunE = nil
		c.PreRunE = nil
		c.Execute()
	})

	return rootCmd

}

func addClusterCmds(cmd *cobra.Command) error {
	// Extract the config file path
	configPath := ""
	if clusterCfg := cmd.Flag("deploy-config"); clusterCfg != nil {
		configPath = clusterCfg.Value.String()
	} else {
		return errstk.New(
			errors.New("deployment config not found"),
			errstk.WithTraceback(),
		)
	}

	cfg, err := config.Parse(configPath)
	if err != nil {
		return err
	}

	switch cfg.Deployment.Environment {
	case model.K3D:
		if k3dClusterCfg, ok := cfg.Deployment.ClusterConfig.(*k3d.ClusterConfig); !ok {
			return errstk.New(
				errors.New("invalid configuration: expected a k3d configuration"),
				errstk.WithTraceback(),
			)
		} else {
			k3dCmd := k3dc.NewK3DClusterCommand(k3dClusterCfg)
			cmd.AddCommand(k3dCmd.Commands()...)
		}
	case model.K3S:
		if k3sClusterCfg, ok := cfg.Deployment.ClusterConfig.(*k3s.ClusterConfig); !ok {
			return errstk.New(
				errors.New("invalid configuration: expected a k3s configuration"),
				errstk.WithTraceback(),
			)
		} else {
			k3sCmd := k3sc.NewK3SClusterCommand(k3sClusterCfg)
			cmd.AddCommand(k3sCmd.Commands()...)
		}
	}

	return nil
}
