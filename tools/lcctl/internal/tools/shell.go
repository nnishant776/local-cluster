package tools

import (
	"log"
	"strings"

	"github.com/nnishant776/local-cluster/config"
	"github.com/spf13/cobra"
)

func NewShellCommand(envConfig map[string]any) *cobra.Command {
	return &cobra.Command{
		Use:  "shell",
		Long: "Run shell commands in the cluster tools environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			return shellCommandHandler(cmd, args)
		},
		DisableFlagParsing: true,
	}
}

func shellCommandHandler(command *cobra.Command, args []string) error {
	ctx := command.Context()
	attachTerm := true
	cmdArgs := []string{"sh"}
	if len(args) > 0 {
		cmdArgs = append(cmdArgs, "-c")
		cmdArgs = append(cmdArgs, strings.Join(args, " "))
		attachTerm = false
	}

	req, err := prepareBaseContainerEnv(config.IMAGE_NAME, cmdArgs)
	if err != nil {
		return err
	}

	if attachTerm {
		req.Config.Tty = true
		req.Config.OpenStdin = true
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
