export projroot:=$(shell realpath .)
export cluster_name:=docker-k3s
export cluster_hostname:=local.cluster.dev

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
	--rm -it \
	-v $$(dirname $$(realpath Makefile)):$$(dirname $$(realpath Makefile))	\
	--security-opt seccomp=unconfined \
	--security-opt label=disable \
	-w $$(dirname $$(realpath Makefile)) \
	golang:latest
build:
ifeq ($(container),true)
	$(eval build: cmd:=)
endif
	$(cmd) go build -v \
		-ldflags "-s -w -X 'github.com/nnishant776/local-cluster/config.k8sVersion=v1.33.1' -X 'k8s.io/component-base/version.gitVersion=v1.33.1' -X 'helm.sh/helm/v4/pkg/chart/v2/util.k8sVersionMinor=33'" \
		-o bin/lcctl github.com/nnishant776/local-cluster


install:
	if [ $$(id -u) != 0 ]; then \
		cp bin/lcctl $$HOME/.local/bin/lcctl; \
	else \
		cp bin/lcctl /usr/local/bin/lcctl; \
	fi
