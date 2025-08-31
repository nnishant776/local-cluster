package config

import (
	"cmp"
	"os"
)

var k8sVersion string = cmp.Or(os.Getenv("K8S_VERSION"), "v1.33.1")

func GetK8SVersion() string {
	return k8sVersion
}
