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

type DeploymentConfig struct {
	Environment      ClusterBackend  `json:"environment,omitempty"`
	RawClusterConfig json.RawMessage `json:"cluster,omitempty"`
	ClusterConfig    any
}

func (self *DeploymentConfig) UnmarshalJSON(b []byte) error {
	type X DeploymentConfig

	dc := X{}
	err := json.Unmarshal(b, &dc)
	if err != nil {
		return err
	}

	switch dc.Environment {
	case K3D:
		dc.ClusterConfig = &k3d.ClusterConfig{}
	case K3S:
		dc.ClusterConfig = &k3s.ClusterConfig{}
	}

	err = json.Unmarshal(dc.RawClusterConfig, dc.ClusterConfig)
	if err != nil {
		return err
	}

	*self = (DeploymentConfig)(dc)

	return nil
}
