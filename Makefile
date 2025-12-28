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
	rm -rf bin && mkdir -p bin && touch bin/.nofile
ifeq ($(container),true)
	$(eval build: cmd:=)
endif
	# Generate basic binary
	env K8S_VERSION=$(k8s_version) go build -v \
		-ldflags "-s -w -X 'github.com/nnishant776/local-cluster/config.k8sVersion=$(k8s_version)'" \
		-o bin/lcctl github.com/nnishant776/local-cluster

	# Generate packaged binary which includes the basic binary
	env K8S_VERSION=$(k8s_version) $(cmd) sh -c \
		"git config --global --add safe.directory $(projroot) && \
		go generate . && \
		cp -r assets/* bin/ && \
		go build -v \
		-ldflags \"-s -w -X 'github.com/nnishant776/local-cluster/config.k8sVersion=$(k8s_version)'\" \
		-o bin/lcctl-$$(uname -s | tr '[:upper:]' '[:lower:]')-$$(uname -m) github.com/nnishant776/local-cluster"


install:
	if [ $$(id -u) != 0 ]; then \
		cp bin/lcctl-$$(uname -s | tr '[:upper:]' '[:lower:]')-$$(uname -m) $$HOME/.local/bin/lcctl; \
	else \
		cp bin/lcctl-$$(uname -s | tr '[:upper:]' '[:lower:]')-$$(uname -m) /usr/local/bin/lcctl; \
	fi
