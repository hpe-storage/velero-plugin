NAME ?= velero-plugin

# The binary to build (prefix).
BIN ?= $(wildcard velero-*)

# This repo's root import path (under GOPATH).
PKG := github.com/hpe-storage/velero-plugin

BUILD_IMAGE ?= golang:1.12-stretch

ifndef REPO_NAME
	REPO_NAME ?= hpestorage/velero-hpe-blockstore
endif

# Use the latest git tag
TAG = $(shell git tag|head -n1)
ifeq ($(TAG),)
	TAG = edge
endif

# unless a BUILD_NUMBER is specified
ifeq ($(IGNORE_BUILD_NUMBER),true)
	VERSION = $(TAG)
else
	ifneq ($(BUILD_NUMBER),)
		VERSION = $(TAG)-$(BUILD_NUMBER)
	else
		VERSION = $(TAG)
	endif
endif

# refers to dockerhub if registry is not specified
IMAGE = $(REPO_NAME):$(VERSION)
ifdef CONTAINER_REGISTRY
	IMAGE = $(CONTAINER_REGISTRY)/$(REPO_NAME):$(VERSION)
endif

# golangci-lint allows us to have a single target that runs multiple linters in
# the same fashion.  This variable controls which linters are used.
LINTER_FLAGS = --disable-all --enable=vet --enable=vetshadow --enable=golint --enable=ineffassign --enable=goconst --enable=deadcode --enable=dupl --enable=varcheck --enable=gocyclo --enable=misspell

# Our target binary is for Linux.  To build an exec for your local (non-linux)
# machine, use go build directly.
ifndef GOOS
	GOOS = linux
endif

GOENV = PATH=$$PATH:$(GOPATH)/bin


.PHONY: help
help:
	@echo "Targets:"
	@echo "    tools    - Download and install go tooling required to build."
	@echo "    vendor   - Download dependencies (go mod vendor)"
	@echo "    compile  - Compiles the source code."
	@echo "    lint     - Static analysis of source code.  Note that this must pass in order to build."
	@echo "    clean    - Remove build artifacts."
	@echo "    image    - Build velero plugin image and create a local docker image.  Errors are ignored."
	@echo "    push     - Push velero plugin image to registry."
	@echo "    all      - Clean, lint, compile, and push image."

.PHONY: check-env
check-env:
ifndef CONTAINER_REGISTRY
	$(error CONTAINER_REGISTRY is undefined)
endif

.PHONY: tools
tools:
	@echo "Get golangci-lint"
	@go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

vendor:
	@go mod vendor

.PHONY: compile
compile:
	@echo "Compiling the source for ${GOOS}"
	@env CGO_ENABLED=0 GOOS=${GOOS} GOARCH=amd64 go build -o build/${NAME} ./velero-hpe-blockstore/main.go

.PHONY: lint
lint:
	@echo "Running lint"
	export $(GOENV) && golangci-lint run $(LINTER_FLAGS) --exclude vendor

# TODO: add tests
.PHONY: test
test:
	@echo "Testing all packages"
	@go test -v ./...

all: clean lint compile image push

image:
	docker build -t $(IMAGE) -f Dockerfile build/

.PHONY: push
push:
	@echo "Publishing velero-plugin:$(VERSION)"
	@docker push $(CONTAINER_REGISTRY)/$(REPO_NAME):$(VERSION)

clean:
	@echo "Removing build artifacts"
	@rm -rf build
	@echo "Removing the image"
	-docker image rm $(CONTAINER_REGISTRY)/$(REPO_NAME):$(VERSION) > /dev/null 2>&1

