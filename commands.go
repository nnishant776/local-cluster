package main

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"dario.cat/mergo"
	errstk "github.com/nnishant776/errstack"
	"github.com/nnishant776/local-cluster/config"
	"github.com/nnishant776/local-cluster/internal/apps"
	k3dc "github.com/nnishant776/local-cluster/internal/cluster/k3d"
	k3sc "github.com/nnishant776/local-cluster/internal/cluster/k3s"
	"github.com/nnishant776/local-cluster/internal/tools"
	"github.com/nnishant776/local-cluster/internal/utils"
	"github.com/nnishant776/local-cluster/pkg/model"
	"github.com/nnishant776/local-cluster/pkg/model/cluster/k3d"
	"github.com/nnishant776/local-cluster/pkg/model/cluster/k3s"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newCLICmds() *cobra.Command {
	rootCmd := rootCmd()
	rootCmd.AddCommand(installCmd())
	rootCmd.AddCommand(uninstallCmd())
	rootCmd.AddCommand(toolsCmd())
	rootCmd.AddCommand(clusterCmd())
	rootCmd.AddCommand(appsCmd())
	return rootCmd
}

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
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

	cmd.PersistentFlags().Bool(
		"verbose", false, "Print verbose errors",
	)

	return cmd
}

func toolsCmd() *cobra.Command {
	toolCmd := &cobra.Command{
		Use:   "tools",
		Short: "A collection of tools to help with cluster management",
		Long:  "A collection of tools to help with cluster management",
	}

	helmCmd := tools.NewHelmCommand()
	helmfileCmd := tools.NewHelmfileCommand()
	kubectlCmd := tools.NewKubectlCommand()
	k9sCmd := tools.NewK9SCommand()

	toolCmd.AddCommand(helmCmd)
	toolCmd.AddCommand(helmfileCmd)
	toolCmd.AddCommand(kubectlCmd)
	toolCmd.AddCommand(k9sCmd)

	return toolCmd
}

func clusterCmd() *cobra.Command {
	cmd := &cobra.Command{
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
			configPath := filepath.Join(utils.GetAppConfigDir(), "config.yaml")

			cfg, _, err := parseConfig(configPath)
			if err != nil {
				return err
			}

			// Append commands based on the cluster type
			switch cfg.Deployment.Environment {
			case model.K3D:
				k3dClusterCfg, ok := cfg.Deployment.ClusterConfig.(*k3d.ClusterConfig)
				if !ok {
					return errstk.NewString(
						"invalid configuration: expected a k3d configuration", errstk.WithStack(),
					)
				}

				cmd.AddCommand(k3dc.NewK3DClusterCommand(k3dClusterCfg).Commands()...)

			case model.K3S:
				k3sClusterCfg, ok := cfg.Deployment.ClusterConfig.(*k3s.ClusterConfig)
				if !ok {
					return errstk.NewString(
						"invalid configuration: expected a k3s configuration", errstk.WithStack(),
					)
				}

				cmd.AddCommand(k3sc.NewK3SClusterCommand(k3sClusterCfg).Commands()...)
			}

			cmd.ParseFlags(args)

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.RunE, cmd.PreRunE = nil, nil
			return cmd.Execute()
		},
	}

	return cmd
}

func appsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "apps",
		Short:            "Manage applications in cluster",
		Long:             "Manage applications in cluster",
		TraverseChildren: true,
	}

	cmd.AddCommand(
		apps.NewInstallCommand(),
		apps.NewUninstallCommand(),
		apps.NewListCommand(),
		apps.NewTemplateCommand(),
	)

	cmd.PersistentFlags().StringP("name", "n", "", "Specify the specific name of the application")

	return cmd
}

func installCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install and setup lcctl",
		Long:  "Install and setup lcctl",
		RunE: func(cmd *cobra.Command, args []string) error {
			configDir := utils.GetAppConfigDir()
			dataDir := utils.GetAppDataDir()
			installDir := utils.GetInstallDir()
			_, err := os.Stat(configDir)
			if err == nil &&
				cmd.Flag("force").Value.String() != "true" {
				return errstk.NewString(
					"config dir exists: setup already done",
					errstk.WithStack(),
				)
			}

			// Clear the existing directory
			if rmErr := os.RemoveAll(configDir); rmErr != nil {
				return errstk.New(rmErr, errstk.WithStack())
			}

			// Copy deployment configuration to correct path
			configSubtree, err := fs.Sub(bundle, "deployment")
			if err != nil {
				return err
			}
			err = os.CopyFS(configDir, configSubtree)
			if err != nil {
				return err
			}

			// Copy cluster binary to correct path
			binSubtree, err := fs.Sub(bundle, "assets")
			if err != nil {
				return err
			}
			binarySourcePath := filepath.Join(dataDir, "bin")
			err = os.CopyFS(binarySourcePath, binSubtree)
			if err != nil {
				return err
			}
			fs.WalkDir(binSubtree, ".", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				if d.Type().IsDir() {
					return nil
				}

				binPath := filepath.Join(binarySourcePath, path)
				if err := os.Chmod(binPath, 0o755); err != nil {
					return err
				}

				return os.Symlink(binPath, filepath.Join(installDir, path))
			})

			configPath := filepath.Join(configDir, "config.yaml")
			k8sVersion := config.GetK8SVersion()
			_, _, err = updateK8SVersionInConfig(k8sVersion, configPath)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().Bool("force", false, "Forcefully perform a fresh install (Existing config will be lost)")

	return cmd
}

func uninstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove lcctl installation",
		Long:  "Remove lcctl installation",
		RunE: func(cmd *cobra.Command, args []string) error {
			configDir := utils.GetAppConfigDir()
			dataDir := utils.GetAppDataDir()
			installDir := utils.GetInstallDir()

			// Clear the existing config directory
			if rmErr := os.RemoveAll(configDir); rmErr != nil {
				return errstk.New(rmErr, errstk.WithStack())
			}

			if cmd.Flag("purge").Value.String() == "true" {
				// Clear the existing data directory
				if rmErr := os.RemoveAll(configDir); rmErr != nil {
					return errstk.New(rmErr, errstk.WithStack())
				}

				// Clear the binaries from the install dir
				binSubtree, err := fs.Sub(bundle, "assets")
				if err != nil {
					return errstk.New(err, errstk.WithStack())
				}
				fs.WalkDir(binSubtree, ".", func(path string, d fs.DirEntry, err error) error {
					if err != nil {
						return err
					}

					if d.Type().IsDir() {
						return nil
					}

					return os.Remove(filepath.Join(installDir, path))
				})
				return os.RemoveAll(dataDir)
			}

			return nil
		},
	}

	cmd.Flags().Bool("purge", false, "Removes the existing copy of the application as well. Otherwise, only configuration is removed")

	return cmd
}

func parseConfig(configPath string) (*model.Config, map[string]any, error) {
	// Open the deployment configuration file
	f, openErr := os.OpenFile(configPath, os.O_RDWR, 0644)
	if openErr != nil {
		return nil, nil, errstk.New(openErr, errstk.WithStack())
	}
	defer f.Close()

	// Parse the config file in a raw map
	rawConfig := map[string]any{}
	if decErr := yaml.NewDecoder(f).Decode(&rawConfig); decErr != nil {
		return nil, nil, errstk.NewChainString(
			"yaml: failed to decode cluster config", errstk.WithStack(),
		).Chain(decErr)
	}

	// Seek to the start of the file again and parse the config file again in the struct
	f.Seek(0, io.SeekStart)
	cfg, parseErr := config.ParseStream(f)
	if parseErr != nil {
		return nil, nil, errstk.NewChainString(
			"cluster: command failed", errstk.WithStack(),
		).Chain(parseErr)
	}

	return cfg, rawConfig, nil
}

func mergeConfig(w io.Writer, rawConfig, cfgOverride map[string]any) error {
	if mErr := mergo.MergeWithOverwrite(&rawConfig, cfgOverride); mErr != nil {
		return errstk.NewChainString(
			"merge: failed to merge k8s config", errstk.WithStack(),
		).Chain(mErr)
	}

	buf := bytes.Buffer{}
	yamlEnc := yaml.NewEncoder(&buf)
	yamlEnc.SetIndent(2)
	if encErr := yamlEnc.Encode(rawConfig); encErr != nil {
		return errstk.NewChainString(
			"yaml: failed to write updated config", errstk.WithStack(),
		).Chain(encErr)
	}

	if _, writeErr := w.Write(buf.Bytes()); writeErr != nil {
		return errstk.New(writeErr, errstk.WithStack())
	}

	return nil
}

func updateK8SVersionInConfig(version string, configPath string) (*model.Config, map[string]any, error) {
	cfg, rawConfig, err := parseConfig(configPath)
	if err != nil {
		return nil, nil, err
	}

	f, openErr := os.OpenFile(configPath, os.O_RDWR, 0644)
	if openErr != nil {
		return nil, nil, errstk.New(openErr, errstk.WithStack())
	}
	defer f.Close()
	cfgOverride := map[string]any{
		"deployment": map[string]any{
			"cluster": map[string]any{
				"k8sVersion": version + "-k3s1",
			},
		},
	}
	if err := mergeConfig(f, rawConfig, cfgOverride); err != nil {
		return nil, nil, err
	}

	return cfg, rawConfig, nil
}
