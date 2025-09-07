export projroot:=$(shell realpath .)
export cluster_name:=docker-k3s
export cluster_hostname:=local.cluster.dev
export k8s_version:=v1.33.2
export k8s_version_major:=1
export k8s_version_minor:=33
export k8s_version_patch:=2

.env:
	echo "PROJECT_ROOT=$(projroot)" >> .env
	echo "CLUSTER_HOSTNAME=$(cluster_hostname)" >> .env
	echo "CLUSTER_NAME=$(cluster_name)" >> .env


.PHONY: build


cert: type:=
cert:
ifeq ($(type), CA)
	export $$(cat .env); . ./scripts/lib.sh; genCA;
endif


build: container:=false
build: runtime:=podman
build: cmd:=$(runtime) run \
	--rm \
	-v $(projroot):$(projroot)	\
	--security-opt seccomp=unconfined \
	--security-opt label=disable \
	-w $(projroot) \
	-e K8S_VERSION=$(k8s_version) \
	golang:latest
build:
ifeq ($(container),true)
	$(eval build: cmd:=)
endif
	env K8S_VERSION=$(k8s_version) $(cmd) sh -c \
		"git config --global --add safe.directory $(projroot) && \
		go generate . && \
		go build -v \
		-ldflags \"-s -w -X 'github.com/nnishant776/local-cluster/config.k8sVersion=$(k8s_version)' -X 'k8s.io/component-base/version.gitVersion=$(k8s_version)' -X 'helm.sh/helm/v4/pkg/chart/v2/util.k8sVersionMinor=$(k8s_version_minor)'\" \
		-o bin/lcctl-$$(uname -s | tr '[:upper:]' '[:lower:]')-$$(uname -m) github.com/nnishant776/local-cluster"


install:
	if [ $$(id -u) != 0 ]; then \
		cp bin/lcctl $$HOME/.local/bin/lcctl; \
	else \
		cp bin/lcctl /usr/local/bin/lcctl; \
	fi
