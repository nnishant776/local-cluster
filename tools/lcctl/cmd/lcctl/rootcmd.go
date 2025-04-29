package main

import (
	"github.com/nnishant776/local-cluster/internal/tools"
	"github.com/spf13/cobra"
)

func newCLICmds() *cobra.Command {
	rootCmd := rootCmd()
	rootCmd.AddCommand(toolsCmd())
	return rootCmd
}

func rootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "lcctl",
		Short: "lcctl is a tool controlling the local cluster deployment",
		Long:  "lcctl is a tool controlling the local cluster deployment",
	}
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
