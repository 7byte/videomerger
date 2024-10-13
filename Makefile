# Default target executed when no arguments are given to make
.DEFAULT_GOAL := all

all: build

build:
	@echo "Building the project..."
	@go build -o bin/ main.go

clean:
	@echo "Cleaning up..."
	@rm -rf bin/

VERSION ?= latest
HTTP_PROXY ?= 
HTTPS_PROXY ?= 
NO_PROXY ?= localhost,127.0.0.1

image:
	@echo "Building the docker image with version $(VERSION)..."
	@docker rmi -f git.7bytes.xyz/merge_xiaomi_monitor_video:$(VERSION) || true
	@docker build --build-arg http_proxy=$(HTTP_PROXY) --build-arg https_proxy=$(HTTPS_PROXY) --build-arg no_proxy=$(NO_PROXY) -t git.7bytes.xyz/merge_xiaomi_monitor_video:$(VERSION) .