package config

import (
	"os"
)

var k8sVersion string

func GetK8SVersion() string {
	if k8sVersion != "" {
		return k8sVersion
	}

	return os.Getenv("K8S_VERSION")
}
