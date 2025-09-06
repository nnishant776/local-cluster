package utils

import (
	"os"
	"path/filepath"
)

func GetAppConfigDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}

	return filepath.Join(configDir, "lcctl")
}
