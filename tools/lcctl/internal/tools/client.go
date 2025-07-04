package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/docker/docker/api/types/backend"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"golang.org/x/term"
)

func createContainerRuntimeClient() (*client.Client, error) {
	client, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)

	return client, err
}

func containerInspect(ctx context.Context, client *client.Client, id string) error {
	log.Printf("Inspecting container")

	inspectResp, err := client.ContainerInspect(ctx, id)
	if err != nil {
		log.Printf("Failed to inspect container: err: %s", err)
		return err
	}

	err = json.NewEncoder(os.Stdout).Encode(inspectResp)
	if err != nil {
		log.Printf("Failed to encode inspect response: err: %s", err)
		return err
	}

	return nil
}

func containerRun(ctx context.Context, client *client.Client, req *backend.ContainerCreateConfig) error {
	// Fetch system information
	sysInfo, infoErr := client.Info(ctx)
	if infoErr != nil {
		return fmt.Errorf("%w: %w", errRuntimeInfoFetch, infoErr)
	}

	// Create the container with the provided spec
	res, createErr := client.ContainerCreate(
		ctx,
		req.Config,
		req.HostConfig,
		req.NetworkingConfig,
		&v1.Platform{Architecture: sysInfo.Architecture, OS: sysInfo.OSType},
		"clustertools",
	)
	if createErr != nil {
		return fmt.Errorf("%w: %w", errContainerCreate, createErr)
	}

	// If tty is requested, put stdin in the terminal mode
	if req.Config.Tty {
		if fd := int(os.Stdin.Fd()); term.IsTerminal(fd) {
			oldState, termErr := term.MakeRaw(fd)
			if termErr != nil {
				return fmt.Errorf("%w: %w", errTerminalConvert, termErr)
			}
			defer term.Restore(fd, oldState)
		}
	}

	// Attach to the created container
	attachResp, attachErr := client.ContainerAttach(ctx, res.ID, container.AttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
		Logs:   false,
	})
	if attachErr != nil {
		return fmt.Errorf("%w: %w", errContainerAttach, attachErr)
	}

	// Start the container
	startErr := client.ContainerStart(ctx, res.ID, container.StartOptions{})
	if startErr != nil {
		return fmt.Errorf("%w: %w", errContainerStart, startErr)
	}

	// Wait for the container to exit
	defer func() {
		waitCh, errCh := client.ContainerWait(ctx, res.ID, container.WaitConditionNextExit)
		select {
		case <-waitCh:
		case <-errCh:
		}
		attachResp.Close()
	}()

	runErr := (error)(nil)
	if req.Config.Tty {
		// If in tty mode, start separate go-routines for routing stdin and stdout
		go func() {
			_, runErr = io.Copy(os.Stdout, attachResp.Reader)
		}()

		go func() {
			_, runErr = io.Copy(attachResp.Conn, os.Stdin)
			if runErr != nil {
				runErr = fmt.Errorf("%w: %w", errContainerInputSend, runErr)
			}
		}()
	} else {
		// If not in tty mode, keep polling the container logs for each stream
		logFile, readErr := client.ContainerLogs(ctx, res.ID, container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
		})
		if readErr != nil {
			runErr = fmt.Errorf("%w: %w", errContainerLogRead, readErr)
		} else {
			_, readErr = stdcopy.StdCopy(os.Stdout, os.Stderr, logFile)
			if readErr != nil {
				runErr = fmt.Errorf("%w: %w", errContainerLogRead, readErr)
			}
		}
	}

	return runErr
}
