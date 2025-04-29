package tools

import (
	"log"

	"github.com/spf13/cobra"
)

func NewHelmfileCommand(envConfig map[string]string) *cobra.Command {
	return &cobra.Command{
		Use:  "helmfile",
		Long: "Run helmfile commands on the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return helmfileCommandHandler(cmd, args)
		},
		DisableFlagParsing: true,
	}
}

func helmfileCommandHandler(command *cobra.Command, args []string) error {
	ctx := command.Context()
	cmdArgs := []string{command.Name()}
	cmdArgs = append(cmdArgs, args...)
	req, err := prepareBaseContainerEnv("ghcr.io/helmfile/helmfile:v0.171.0", cmdArgs)
	if err != nil {
		return err
	}

	client, err := createContainerRuntimeClient()
	if err != nil {
		return err
	}

	err = containerRun(ctx, client, req)
	if err != nil {
		log.Printf("Failed to run container: err: %s", err)
	}

	return nil
}
