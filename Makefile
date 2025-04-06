export projroot:=$(shell realpath .)
export cluster_name:=docker-k3s
export cluster_hostname:=local.cluster.dev
export registry_port:=5000
export k3d_version:=v5.8.3
export kubectl_version:=v1.32.3
export k3s_version:=v1.30.6-k3s1
export k9s_version:=v0.40.10
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
		curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | TAG=$(k3d_version) bash ; \
	fi
	if ! command -v kubectl; then \
		. ./lib.sh; curl -fSLGO "https://dl.k8s.io/release/$(kubectl_version)/bin/linux/$$(cpuArch)/kubectl" ; \
		chmod 0755 kubectl; sudo mv kubectl /usr/local/bin ; \
	fi
	if ! command -v k9s; then \
		. ./lib.sh; curl -fSLGO https://github.com/derailed/k9s/releases/download/$(k9s_version)/k9s_$$(uname)_$$(cpuArch).tar.gz ; \
		tar -xf k9s_$$(uname)_$$(cpuArch).tar.gz; chmod 0755 k9s; sudo mv k9s /usr/local/bin; rm k9s_$$(uname)_$$(cpuArch).tar.gz ; \
	fi


cluster: action:=
cluster: setup
ifeq ($(action), create)
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
		-v $$HOME/.kube/config:/helm/.kube/config:ro \
		-it \
		-w $(projroot) \
		ghcr.io/helmfile/helmfile:v0.171.0 \
		helmfile sync -f $(projroot)/addons/helmfile.yaml
else ifeq ($(action), uninstall)
	docker run \
		--rm \
		--env-file .env \
		--network host \
		--security-opt seccomp=unconfined \
		--security-opt label=disable \
		-v $(projroot):$(projroot):ro \
		-v $$HOME:$$HOME:ro \
		-v $$HOME/.kube/config:/helm/.kube/config:ro \
		-it \
		-w $(projroot) \
		ghcr.io/helmfile/helmfile:v0.171.0 \
		helmfile destroy -f $(projroot)/addons/helmfile.yaml
else ifeq ($(action), debug)
	docker run \
		--rm \
		--env-file .env \
		--network host \
		--security-opt seccomp=unconfined \
		--security-opt label=disable \
		-v $(projroot):$(projroot):ro \
		-v $$HOME:$$HOME:ro \
		-v $$HOME/.kube/config:/helm/.kube/config:ro \
		-it \
		-w $(projroot) \
		ghcr.io/helmfile/helmfile:v0.171.0 \
		helmfile template -f $(projroot)/addons/helmfile.yaml --debug
else
	echo "Unknown action. Exiting"
endif

cert: type:=
cert:
ifeq ($(type), CA)
	export $$(cat .env); . ./lib.sh; genCA;
endif
