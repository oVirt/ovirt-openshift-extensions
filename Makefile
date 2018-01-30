# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GODEP=dep

FLEX_DRIVER_BINARY_NAME=ovirt-flexdriver
PROVISIONER_BINARY_NAME=ovirt-provisioner

IMAGE=rgolangh/ovirt-provisioner
VERSION?=$(shell git describe --tags --always | cut -d "-" -f1)
COMMIT=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

COMMON_ENV=CGO_ENABLED=0 GOOS=linux GOARCH=amd64
COMMON_GO_BUILD_FLAGS=-a -ldflags '-extldflags "-static"'

all: clean deps build test

build-flex:
	$(COMMON_ENV) $(GOBUILD) \
	$(COMMON_GO_BUILD_FLAGS) \
	-o $(FLEX_DRIVER_BINARY_NAME) \
	-v cmd/$(FLEX_DRIVER_BINARY_NAME)/*.go

build-provisioner:
	$(COMMON_ENV) $(GOBUILD) \
	$(COMMON_GO_BUILD_FLAGS) \
	-o $(PROVISIONER_BINARY_NAME) \
	-v cmd/$(PROVISIONER_BINARY_NAME)/*.go

container:
	build
	quick-container

quick-container:
	cp $(PROVISIONER_BINARY_NAME) deployment/container
	docker build -t $(IMAGE):$(VERSION) deployment/container/

push:
    # don't forget docker login. TODO official registry
	docker push $(IMAGE):$(VERSION)

build: \
	build-flex \
	build-provisioner

test:
	$(GOTEST) -v ./...
clean:
	$(GOCLEAN)
	rm -f $(FLEX_DRIVER_BINARY_NAME)
	rm -f $(PROVISIONER_BINARY_NAME)
run: \
	build \
	./$(FLEX_DRIVER_BINARY_NAME)
	./$(PROVISIONER_BINARY_NAME)
deps:
	glide install --strip-vendor

.PHONY: build-flex build-provisioner