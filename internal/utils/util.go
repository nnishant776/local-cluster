package utils

import (
	"fmt"
	"os"
	"path/filepath"
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
