.PHONY: build

build:
	go build -v -ldflags "-s -w" -o bin/lcctl github.com/nnishant776/local-cluster/cmd/lcctl

install:
	if [ $$(id -u) != 0 ]; then \
		cp bin/lcctl $$HOME/.local/bin/lcctl; \
	else \
		cp bin/lcctl /usr/local/bin/lcctl; \
	fi
