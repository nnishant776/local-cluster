package common

type PathMapping struct {
	Host    string `json:"host,omitempty"`
	Cluster string `json:"cluster,omitempty"`
}

type BuiltinServices struct {
	LoadBalancer       bool `json:"loadBalancer,omitempty"`
	IngressController  bool `json:"ingressController,omitempty"`
	MetricsServer      bool `json:"metricsServer,omitempty"`
	StorageProvisioner bool `json:"storageProvisioner,omitempty"`
}

type BootstrapConfig struct {
	DNS bool `json:"dns,omitempty"`
}

type RegistryConfig struct {
	Enabled bool           `json:"enabled,omitempty"`
	Mirror  RegistryMirror `json:"mirror,omitempty"`
}

type RegistryMirror struct {
	Endpoints []RegistryEndpoint `json:"endpoints,omitempty"`
}

type RegistryEndpoint struct {
	Host string           `json:"host,omitempty"`
	Auth RegistryAuthInfo `json:"auth,omitempty"`
	TLS  RegistryTLSInfo  `json:"tls,omitempty"`
}

type RegistryAuthInfo struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type RegistryTLSInfo struct {
	Enabled  bool   `json:"enabled,omitempty"`
	CertFile string `json:"cert_file,omitempty"`
	KeyFile  string `json:"key_file,omitempty"`
	CAFile   string `json:"ca_file,omitempty"`
}

type BaseClusterConfig struct {
	Name            string          `json:"name,omitempty"`
	K8SVersion      string          `json:"k8sVersion,omitempty"`
	DataPath        PathMapping     `json:"dataPath,omitempty"`
	VolumeMounts    []PathMapping   `json:"volumeMounts,omitempty"`
	BuiltinServcies BuiltinServices `json:"services,omitempty"`
	Bootstrap       BootstrapConfig `json:"bootstrap,omitempty"`
	Registry        RegistryConfig  `json:"registry,omitempty"`
}
