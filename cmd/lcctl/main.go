package main

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io"
	"os"
	"os/signal"

	"dario.cat/mergo"
	"github.com/nnishant776/local-cluster/config"
	"github.com/nnishant776/local-cluster/internal/apps"
	k3dc "github.com/nnishant776/local-cluster/internal/cluster/k3d"
	k3sc "github.com/nnishant776/local-cluster/internal/cluster/k3s"
	"github.com/nnishant776/local-cluster/internal/tools"
	"github.com/nnishant776/local-cluster/pkg/model"
	"github.com/nnishant776/local-cluster/pkg/model/cluster/k3d"
	"github.com/nnishant776/local-cluster/pkg/model/cluster/k3s"
	"gopkg.in/yaml.v3"

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
	rootCmd.AddCommand(appsCmd())
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
		Use:   "tools",
		Short: "A collection of tools to help with cluster management",
		Long:  "A collection of tools to help with cluster management",
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

			rawConfig, cfg := map[string]any{}, (*model.Config)(nil)

			if f, err := os.Open(configPath); err != nil {
				return errstk.New(err, errstk.WithStack())
			} else {
				defer f.Close()
				err = yaml.NewDecoder(f).Decode(&rawConfig)
				if err != nil {
					return errstk.NewChainString(
						"yaml: failed to decode cluster config", errstk.WithStack(),
					).Chain(err)
				}

				f.Seek(0, io.SeekStart)
				cfg, err = config.ParseStream(f)
				if err != nil {
					return errstk.NewChainString(
						"cluster: command failed", errstk.WithStack(),
					).Chain(err)
				}
			}

			switch cfg.Deployment.Environment {
			case model.K3D:
				if k3dClusterCfg, ok := cfg.Deployment.ClusterConfig.(*k3d.ClusterConfig); ok {
					cmd.AddCommand(k3dc.NewK3DClusterCommand(k3dClusterCfg).Commands()...)
					if err := mergo.MergeWithOverwrite(
						&rawConfig,
						map[string]any{
							"deployment": map[string]any{
								"cluster": map[string]any{
									"k8sVersion": config.K8S_VERSION + "-k3s1",
								},
							},
						},
					); err != nil {
						return errstk.NewChainString(
							"merge: failed to merge k8s config", errstk.WithStack(),
						).Chain(err)
					} else {
						buf := bytes.Buffer{}
						yamlEnc := yaml.NewEncoder(&buf)
						yamlEnc.SetIndent(2)
						if err := yamlEnc.Encode(rawConfig); err != nil {
							return errstk.NewChainString(
								"yaml: failed to write updated config", errstk.WithStack(),
							).Chain(err)
						} else {
							err = os.WriteFile(configPath, buf.Bytes(), 0o644)
							if err != nil {
								return errstk.New(err, errstk.WithStack())
							}
						}
					}
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

			cmd.ParseFlags(args)

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

func appsCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:              "apps",
		Short:            "Manage applications in cluster",
		Long:             "Manage applications in cluster",
		TraverseChildren: true,
	}

	rootCmd.AddCommand(
		apps.NewInstallCommand(),
		apps.NewUninstallCommand(),
		apps.NewListCommand(),
		apps.NewTemplateCommand(),
	)

	rootCmd.PersistentFlags().StringP("name", "n", "", "Specify the specific name of the application")

	return rootCmd
}

func main() {
	ctx, cancelFn := signal.NotifyContext(context.Background(), []os.Signal{unix.SIGTERM, unix.SIGINT}...)
	defer cancelFn()

	cmd := newCLICmds()
	switch os.Getenv("TOOL_MODE") {
	case "helm":
		cmd = tools.NewHelmCommand(nil)
	case "helmfile":
		cmd = tools.NewHelmfileCommand(nil)
	case "k9s":
		cmd = tools.NewK9SCommand(nil)
	case "kubectl":
		cmd = tools.NewKubectlCommand(nil)
	}

	err := cmd.ExecuteContext(ctx)
	if flg := cmd.Flag("verbose"); err != nil && flg != nil && flg.Value.String() == "true" {
		fmt.Printf("Failed to execute command: %#4v\n", err)
	}
}
