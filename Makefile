BINARY_NAME=apartapatia-runtime
ALPINE_VERSION ?= v3.23
ALPINE_PATCH_VERSION ?= 3.23.3

.PHONY: all build run clean setup-rootfs test

all: build

setup-rootfs:
	@echo "скачивание и распаковка rootfs $(ALPINE_PATCH_VERSION)..."
	mkdir -p rootfs
	@if [ ! -f rootfs/bin/sh ]; then \
		echo "...скачивание alpine..."; \
		curl -fL -o rootfs/alpine.tar.gz http://dl-cdn.alpinelinux.org/alpine/$(ALPINE_VERSION)/releases/x86_64/alpine-minirootfs-$(ALPINE_PATCH_VERSION)-x86_64.tar.gz; \
		echo "...распаковка архива..."; \
		tar -xf rootfs/alpine.tar.gz -C rootfs; \
		echo "...очистка временных файлов..."; \
		rm rootfs/alpine.tar.gz; \
		echo "...готово!..."; \
	else \
		echo "...rootfs уже существует, скачивание пропущено..."; \
	fi

build: 
	go build -o $(BINARY_NAME) main.go
			
run: build setup-rootfs
	sudo ./$(BINARY_NAME) run

test: build setup-rootfs
	sudo ./$(BINARY_NAME) run < /dev/null

clean:
	rm -f $(BINARY_NAME)
	rm -rf rootfs
