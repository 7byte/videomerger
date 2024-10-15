# Default target executed when no arguments are given to make
.DEFAULT_GOAL := all

all: build

build:
	@echo "Building the project..."
	@go build -o bin/ main.go

clean:
	@echo "Cleaning up..."
	@rm -rf bin/ images/

VERSION ?= latest
IMAGE_NAME = git.7bytes.xyz/merge_xiaomi_monitor_video:$(VERSION)

ARCH := $(shell uname -m)
PLATFORM ?= linux/$(ARCH)

HTTP_PROXY ?= 
HTTPS_PROXY ?= 
NO_PROXY ?= localhost,127.0.0.1

docker_builder:
	@echo "Create the docker builder..."
	docker buildx inspect multi-builder > /dev/null || docker buildx create --driver docker-container --platform linux/amd64,linux/arm64 --name multi-builder --use
	docker buildx inspect --bootstrap

image:
	@echo "Building the docker image with version $(VERSION)..."
	@docker images|grep $(IMAGE_NAME) && docker rmi $(IMAGE_NAME) || true
	@mkdir -p images
	docker buildx build --platform $(PLATFORM) --build-arg http_proxy=$(HTTP_PROXY) --build-arg https_proxy=$(HTTPS_PROXY) --build-arg no_proxy=$(NO_PROXY) -t $(IMAGE_NAME) -o type=docker .