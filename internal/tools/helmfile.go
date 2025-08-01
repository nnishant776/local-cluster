package tools

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/nnishant776/errstack"
	"github.com/spf13/cobra"

	"github.com/helmfile/helmfile/cmd"
	"github.com/helmfile/helmfile/pkg/app"
	helmfilecfg "github.com/helmfile/helmfile/pkg/config"
	helmfileerr "github.com/helmfile/helmfile/pkg/errors"
)

func NewHelmfileCommand(envConfig map[string]any) *cobra.Command {
	return &cobra.Command{
		Use:                "helmfile",
		Long:               "Run helmfile commands on the cluster",
		DisableFlagParsing: true,
		RunE: func(c *cobra.Command, args []string) error {
			sigs, errChan := make(chan os.Signal, 1), make(chan error, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				rootCmd, err := cmd.NewRootCmd(&helmfilecfg.GlobalOptions{})
				if err != nil {
					errChan <- errstack.NewChainString(
						"helmfile: failed to instantiate", errstack.WithStack(),
					).Chain(err)

					return
				}

				rootCmd.SetArgs(args)

				errChan <- rootCmd.Execute()
			}()

			select {
			case sig := <-sigs:
				if sig != nil {
					app.Cancel()
					app.CleanWaitGroup.Wait()

					// See http://tldp.org/LDP/abs/html/exitcodes.html
					switch sig {
					case syscall.SIGINT:
						os.Exit(130)
					case syscall.SIGTERM:
						os.Exit(143)
					}
				}
			case err := <-errChan:
				defer helmfileerr.HandleExitCoder(err)
				return err
			}

			return nil
		},
	}
}
