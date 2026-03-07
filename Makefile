export projroot:=$(shell realpath .)
export k8s_version:=v1.35.2
export k8s_version_major:=1
export k8s_version_minor:=35
export k8s_version_patch:=2

.PHONY: build


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
	rm -rf bin && mkdir -p bin
ifeq ($(container),true)
	$(eval build: cmd:=)
endif
	# Generate basic binary
	env K8S_VERSION=$(k8s_version) $(cmd) sh -c \
		"go build -v \
		-ldflags \"-s -w -X 'github.com/nnishant776/local-cluster/config.k8sVersion=$(k8s_version)'\" \
		-o bin/lcctl github.com/nnishant776/local-cluster"

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
