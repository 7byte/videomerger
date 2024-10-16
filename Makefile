# Default target executed when no arguments are given to make
.DEFAULT_GOAL := all

.PHONY: all build clean multi_builder image

all: build

build:
	@echo "Building the project..."
	@go build -o bin/videomerger main.go

clean:
	@echo "Cleaning up..."
	@rm -rf bin/ images/

VERSION ?= latest
IMAGE_NAME = github.com/7byte/videomerger:$(VERSION)

ARCH := $(shell uname -m)
PLATFORM ?= linux/$(ARCH)

HTTP_PROXY ?= 
HTTPS_PROXY ?= 
NO_PROXY ?= localhost,127.0.0.1

multi_builder:
	@echo "Create the docker builder..."
	docker buildx inspect multi-builder > /dev/null || docker buildx create --driver docker-container --platform linux/amd64,linux/arm64 --name multi-builder --use
	docker buildx inspect --bootstrap

image:
	@echo "Building the docker image with version $(VERSION)..."
	@docker images|grep $(IMAGE_NAME) && docker rmi $(IMAGE_NAME) || true
	docker buildx build --platform $(PLATFORM) --build-arg http_proxy=$(HTTP_PROXY) --build-arg https_proxy=$(HTTPS_PROXY) --build-arg no_proxy=$(NO_PROXY) -t $(IMAGE_NAME) -o type=docker .