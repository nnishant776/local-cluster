package config

const (
	APP_VERSION string = "v1.0.1"
	IMAGE_NAME  string = "k8s-tools" + ":" + APP_VERSION
)

var k8sVersion string = "v1.33.1"

func GetK8SVersion() string {
	return k8sVersion
}
