package tools

import (
	"os"

	"github.com/derailed/k9s/cmd"
	"github.com/spf13/cobra"
)

func NewK9SCommand(envConfig map[string]any) *cobra.Command {
	return &cobra.Command{
		Use:                "k9s",
		Long:               "Run k9s",
		DisableFlagParsing: true,
		RunE: func(c *cobra.Command, args []string) error {
			os.Args = append([]string{c.Name()}, args...)
			cmd.Execute()
			return nil
		},
	}
}
