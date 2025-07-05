package k3d

import "github.com/nnishant776/local-cluster/pkg/model/cluster/common"

type NetworkConfig struct {
	Subnet string `json:"subnet,omitempty"`
}

type WorkloadConfig struct {
	Servers int `json:"server,omitempty"`
	Workers int `json:"worker,omitempty"`
}

type APIServerConfig struct {
	IP   string `json:"ip,omitempty"`
	Port string `json:"port,omitempty"`
}

type ClusterConfig struct {
	common.BaseClusterConfig
	Network   NetworkConfig   `json:"network,omitempty"`
	ApiServer APIServerConfig `json:"apiServer,omitempty"`
}
