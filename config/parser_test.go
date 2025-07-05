package config

import (
	"strings"
	"testing"

	"github.com/nnishant776/local-cluster/pkg/model/cluster/k3d"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	var testConfig = `
infra:
  rootDomain: &san local.cluster.dev
  tls:
    privateCA:
      countryCode: UC
      state: Andromeda
      locality: Mirach
      commonName: universe.com
      organization: Constellation
      organizationUnit: Development
      emailAddress: mail@universe.com
deployment:
  environment: k3d     # k3d, k3s
  cluster:             # Values for cluster should reflect based on environment
    name: docker-k3s
    k8sVersion: v1.30.6-k3s1
    dataPath:
      host: /mnt
      cluster: /mnt
    volumeMounts:
      - host: ${PROJECT_ROOT}/.registry
        cluster: /var/lib/registry
    services:
      loadBalancer: false
      ingressController: false
      metricsServer: false
      storageProvisioner: false
    bootstrap:
      dns: true
    registry:
      enabled: true
      mirror:
        endpoints: []
          # - host: "registry.local.cluster.dev:5000"
          #   auth:
          #     username:  # username
          #     password:  # password
          #   tls:
          #     enabled: false
          #     cert_file: # path to the cert file used in the registry
          #     key_file:  # path to the key file used in the registry
          #     ca_file:   # path to the ca file used in the registry
tools:
  k3d:
    install: true
    version: v5.8.3
  kubectl:
    install: true
    version: v1.32.3
  k9s:
    install: true
    version: v0.50.4
`

	cfg, err := ParseStream(strings.NewReader(testConfig))
	assert.NoError(t, err)
	assert.Equal(t, "k3d", cfg.Deployment.Environment.String())
	k3dCfg, ok := cfg.Deployment.ClusterConfig.(*k3d.ClusterConfig)
	assert.True(t, true, ok)
	assert.Equal(t, "docker-k3s", k3dCfg.Name)
	assert.Equal(t, true, k3dCfg.Bootstrap.DNS)
}
