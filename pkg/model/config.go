package model

import (
	"encoding/json"

	"github.com/nnishant776/local-cluster/pkg/model/cluster/k3d"
	"github.com/nnishant776/local-cluster/pkg/model/cluster/k3s"
)

type ClusterBackend string

func (self ClusterBackend) String() string {
	return string(self)
}

const (
	K3D ClusterBackend = "k3d"
	K3S ClusterBackend = "k3s"
)

type Config struct {
	Deployment DeploymentConfig `json:"deployment,omitempty"`
}

type _DeploymentConfig struct {
	Environment      ClusterBackend  `json:"environment,omitempty"`
	RawClusterConfig json.RawMessage `json:"cluster,omitempty"`
}

type DeploymentConfig struct {
	Environment   ClusterBackend `json:"environment,omitempty"`
	ClusterConfig any
}

type ToolsConfig struct {
	Image string      `json:"image,omitempty"`
	Apps  []AppConfig `json:"apps,omitempty"`
}

type AppConfig struct {
	Name      string `json:"name,omitempty"`
	Installed bool   `json:"install,omitempty"`
	Version   string `json:"version,omitempty"`
}

func (self *DeploymentConfig) UnmarshalJSON(b []byte) error {
	dc := _DeploymentConfig{}
	err := json.Unmarshal(b, &dc)
	if err != nil {
		return err
	}

	self.Environment = dc.Environment

	switch dc.Environment {
	case K3D:
		self.ClusterConfig = &k3d.ClusterConfig{}
	case K3S:
		self.ClusterConfig = &k3s.ClusterConfig{}
	}

	err = json.Unmarshal(dc.RawClusterConfig, self.ClusterConfig)
	if err != nil {
		return err
	}

	return nil
}
