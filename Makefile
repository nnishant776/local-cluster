export projroot:=$(shell realpath .)
export cluster_name:=docker-k3s
export cluster_hostname:=local.cluster.dev
export registry_port:=5000
export k3d_version:=v5.8.3
export k3s_version:=v1.30.6-k3s1
export data_path_src:=/mnt/$(cluster_name)
export data_path_dest:=/mnt

.env:
	echo "PROJECT_ROOT=$(projroot)" >> .env
	echo "CLUSTER_HOSTNAME=$(cluster_hostname)" >> .env
	echo "CLUSTER_NAME=$(cluster_name)" >> .env
	echo "REGISTRY_PORT=$(registry_port)" >> .env
	echo "K3S_VERSION=$(k3s_version)" >> .env
	echo "DATA_PATH_SRC=$(data_path_src)" >> .env
	echo "DATA_PATH_DEST=$(data_path_dest)" >> .env

setup: .env
	if ! command -v k3d; then \
		curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | TAG=$(k3d_version) bash; \
	fi

cluster: action:=
cluster: setup
ifeq ($(action), create)
	export $$(cat .env) ORIGIN='$$ORIGIN' TTL='$$TTL'; cd cluster/bootstrap; cat coredns-custom.yaml.in | envsubst | tee coredns-custom.yaml 2>&1 > /dev/null
	export $$(cat .env); cd cluster; cat config.yaml | envsubst | k3d cluster create --config -
else ifeq ($(action), destroy)
	k3d cluster delete $(cluster_name)
else
	echo "Unknown action. Exiting"
endif

addons: actions:=
addons: setup
ifeq ($(action), install)
	docker run \
		--rm \
		--env-file .env \
		--network host \
		--security-opt seccomp=unconfined \
		--security-opt label=disable \
		-v $(projroot):$(projroot):ro \
		-v $$HOME:$$HOME:ro \
		-v $$HOME/.kube/config:/root/.kube/config:ro \
		-it \
		-w $(projroot) \
		ghcr.io/helmfile/helmfile:v0.171.0 \
		helmfile --kubeconfig $$HOME/.kube/config sync -f $(projroot)/addons/helmfile.yaml
else ifeq ($(action), uninstall)
	docker run \
		--rm \
		--env-file .env \
		--network host \
		--security-opt seccomp=unconfined \
		--security-opt label=disable \
		-v $(projroot):$(projroot):ro \
		-v $$HOME:$$HOME:ro \
		-v $$HOME/.kube/config:/root/.kube/config:ro \
		-it \
		-w $(projroot) \
		ghcr.io/helmfile/helmfile:v0.171.0 \
		helmfile --kubeconfig $$HOME/.kube/config destroy -f $(projroot)/addons/helmfile.yaml
else ifeq ($(action), debug)
	docker run \
		--rm \
		--env-file .env \
		--network host \
		--security-opt seccomp=unconfined \
		--security-opt label=disable \
		-v $(projroot):$(projroot):ro \
		-v $$HOME:$$HOME:ro \
		-v $$HOME/.kube/config:/root/.kube/config:ro \
		-it \
		-w $(projroot) \
		ghcr.io/helmfile/helmfile:v0.171.0 \
		helmfile --kubeconfig $$HOME/.kube/config template -f $(projroot)/addons/helmfile.yaml --debug
else
	echo "Unknown action. Exiting"
endif

cert: type:=
cert:
ifeq ($(type), CA)
	export $$(cat .env); . ./certgen; genCA;
endif
