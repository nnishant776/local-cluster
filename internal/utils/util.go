package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	errstk "github.com/nnishant776/errstack"
	"github.com/nnishant776/local-cluster/config"
	"github.com/nnishant776/local-cluster/pkg/model"
	"gopkg.in/yaml.v3"
)

func GetAppConfigDir() string {
	configDir := "/etc"

	if os.Getuid() != 0 {
		cfgDir, err := os.UserConfigDir()
		if err != nil {
			panic(err)
		}

		configDir = cfgDir
	}

	return filepath.Join(configDir, "lcctl")
}

func GetInstallDir() string {
	installDir := "/usr/local/bin"

	if os.Getuid() != 0 {
		instDir, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}

		installDir = filepath.Join(instDir, ".local", "bin")
	}

	return installDir
}

func GetAppDataDir() string {
	dataDir := "/usr/local/share"

	if os.Getuid() != 0 {
		dtDir, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}

		dataDir = filepath.Join(dtDir, ".local", "share")
	}

	return filepath.Join(dataDir, "lcctl")
}

func GetAppRuntimeDir() string {
	runtimeDir := "/run/lcctl"

	if os.Getuid() != 0 {
		runtimeDir = fmt.Sprintf("/run/user/%d/lcctl", os.Getuid())
	}

	return runtimeDir
}

func ParseConfig(configPath string) (*model.Config, map[string]any, error) {
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
