BINARY		= apartapatia-runtime
ALPINE_V	= v3.23
ALPINE_PV	= 3.23.3
LINT_V		= 2.11.4
LINT_VR 	= 2

.PHONY: all build run test clean setup-rootfs

all: build

build:
	go build -o $(BINARY) main.go

setup-rootfs:
	@mkdir -p rootfs
	@[ -f rootfs/bin/sh ] || ( \
		curl -fL -o rootfs/alpine.tar.gz \
		  http://dl-cdn.alpinelinux.org/alpine/$(ALPINE_V)/releases/x86_64/alpine-minirootfs-$(ALPINE_PV)-x86_64.tar.gz \
		&& tar -xf rootfs/alpine.tar.gz -C rootfs \
		&& rm rootfs/alpine.tar.gz )

run: build setup-rootfs
	sudo ./$(BINARY) run

test: build setup-rootfs
	go run github.com/golangci/golangci-lint/v$(LINT_VR)/cmd@v$(LINT_V) run ./... || true
	go test -c -o test_runner .
	go run gotest.tools/gotestsum@latest \
		--format testname \
		--junitfile test-results.xml \
		--raw-command -- \
		bash -c "sudo ./test_runner -test.v -test.count=1 | go tool test2json -t -p $(BINARY)"

clean:
	rm -f $(BINARY) test_runner test-results.*
	rm -rf rootfs
