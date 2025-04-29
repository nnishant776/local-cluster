package tools

import (
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
)

func prepareBaseContainerEnv(image string, command []string) (*container.CreateRequest, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	config := &container.Config{
		// AttachStdin:  true,
		// AttachStdout: true,
		// AttachStderr: true,
		// Tty:          true,
		// OpenStdin:    true,
		Image:      image,
		Entrypoint: strslice.StrSlice(command),
		WorkingDir: wd,
	}

	hostConfig := &container.HostConfig{
		// CapAdd:      []string{},
		// CapDrop:     []string{"ALL"},
		NetworkMode: "host",
		AutoRemove:  true,
		Privileged:  false,
		SecurityOpt: []string{
			"seccomp=unconfined",
			"label=disable",
		},
		Mounts: []mount.Mount{
			{
				Type:     "bind",
				Source:   os.ExpandEnv("$HOME"),
				Target:   os.ExpandEnv("$HOME"),
				ReadOnly: true,
			},
			{
				Type:     "bind",
				Source:   os.ExpandEnv("$HOME/.kube/config"),
				Target:   "/helm/.kube/config",
				ReadOnly: true,
			},
			{
				Type:   "bind",
				Source: os.ExpandEnv("$HOME/.cache/helm"),
				Target: "/helm/.cache/helm",
			},
			{
				Type:   "bind",
				Source: os.ExpandEnv("$HOME/.config/helm"),
				Target: "/helm/.config/helm",
			},
			{
				Type:   "bind",
				Source: os.ExpandEnv("$HOME/.local/share/helm"),
				Target: "/helm/.local/share/helm",
			},
			{
				Type:   "bind",
				Source: wd,
				Target: wd,
			},
		},
	}

	return &container.CreateRequest{
		Config:     config,
		HostConfig: hostConfig,
	}, nil
}
