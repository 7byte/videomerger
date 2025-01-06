# Default target executed when no arguments are given to make
.DEFAULT_GOAL := all

.PHONY: all build clean multi_builder image

all: build

# install opencv: https://gocv.io/getting-started/
build:
	@echo "Building the project..."
	@go build -o bin/videomerger main.go

clean:
	@echo "Cleaning up..."
	@rm -rf bin/ images/ dist/

VERSION ?= latest
IMAGE_NAME = github.com/7byte/videomerger:$(VERSION)
IMAGE_NAME_FAT = $(IMAGE_NAME)-fat

ARCH := $(shell uname -m)
PLATFORM ?= linux/amd64

HTTP_PROXY ?= 
HTTPS_PROXY ?= 
NO_PROXY ?= localhost,127.0.0.1

multi_builder:
	@echo "Create the docker builder..."
	docker buildx inspect multi-builder > /dev/null || docker buildx create --driver docker-container --platform linux/amd64,linux/arm64 --name multi-builder --use
	docker buildx inspect --bootstrap

# image:
# 	@echo "Building the docker image with version $(VERSION)..."
# 	@docker inspect $(IMAGE_NAME) > /dev/null && docker rmi $(IMAGE_NAME) || true
# 	docker buildx build --platform $(PLATFORM) --build-arg http_proxy=$(HTTP_PROXY) --build-arg https_proxy=$(HTTPS_PROXY) --build-arg no_proxy=$(NO_PROXY) -t $(IMAGE_NAME) -o type=docker .

image:
	@echo "Building the docker image with version $(VERSION)..."
	@docker inspect $(IMAGE_NAME) > /dev/null && docker rmi $(IMAGE_NAME) || true
	docker buildx build --platform $(PLATFORM) --build-arg http_proxy=$(HTTP_PROXY) --build-arg https_proxy=$(HTTPS_PROXY) --build-arg no_proxy=$(NO_PROXY) -f Dockerfile -t $(IMAGE_NAME_FAT) -o type=docker .
	slim build --http-probe=false --continue-after=1 --include-bin /usr/bin/ffmpeg --tag $(IMAGE_NAME) $(IMAGE_NAME_FAT)
	docker rmi $(IMAGE_NAME_FAT)
