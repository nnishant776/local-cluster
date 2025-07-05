package tools

import (
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"

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
			sig, sigs, errChan := (os.Signal)(nil), make(chan os.Signal, 1), make(chan error, 1)

			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
			rootCmd, err := cmd.NewRootCmd(&helmfilecfg.GlobalOptions{})
			rootCmd.SetArgs(args)

			go func() {
				if err != nil {
					errChan <- err
					return
				}

				errChan <- rootCmd.Execute()
			}()

			select {
			case sig = <-sigs:
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
