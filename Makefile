export projroot:=$(shell realpath .)
export cluster_hostname:=local.cluster.dev
export registry_port:=5000
export k3s_version:=v1.30.6-k3s1

.env:
	echo "PROJECT_ROOT=$(projroot)" >> .env
	echo "CLUSTER_HOSTNAME=$(cluster_hostname)" >> .env
	echo "REGISTRY_PORT=$(registry_port)" >> .env
	echo "K3S_VERSION=$(k3s_version)" >> .env

setup: .env

cluster: action:=
cluster: setup
ifeq ($(action), create)
	export $$(cat .env); cd cluster; cat config.yaml | envsubst | k3d cluster create --config -
else ifeq ($(action), destroy)
	k3d cluster delete
else
	echo "Unknown action. Exiting"
endif
