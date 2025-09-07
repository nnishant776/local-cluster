package k3d

import (
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	errstk "github.com/nnishant776/errstack"
	"github.com/nnishant776/local-cluster/internal/utils"
	"github.com/nnishant776/local-cluster/pkg/model/cluster/k3s"
	"github.com/spf13/cobra"
)

func NewStopCommand(cfg *k3s.ClusterConfig) *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the cluster",
		Long:  "Stop the cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Read the PID file
			pidFilePath := filepath.Join(utils.GetAppRuntimeDir(), "pid")
			pidStr, err := os.ReadFile(pidFilePath)
			if err != nil {
				return errstk.New(err, errstk.WithStack())
			}

			pid, err := strconv.ParseInt(string(pidStr), 10, 32)
			if err != nil {
				return errstk.New(err, errstk.WithStack())
			}

			if pid <= 1 {
				return errstk.NewString(
					"invalid cluster process id: "+string(pidStr), errstk.WithStack(),
				)
			}

			proc, err := os.FindProcess(int(pid))
			if err != nil {
				return errstk.New(err, errstk.WithStack())
			}

			if err := proc.Signal(syscall.SIGTERM); err != nil {
				return errstk.New(err, errstk.WithStack())
			}

			if err := os.RemoveAll(pidFilePath); err != nil {
				return errstk.New(err, errstk.WithStack())
			}

			return nil
		},
	}

	return startCmd
}
