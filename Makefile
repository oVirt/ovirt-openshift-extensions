# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GODEP=dep

PREFIX=.
ARTIFACT_DIR ?= .

FLEX_DRIVER_BINARY_NAME=ovirt-flexvolume-driver
PROVISIONER_BINARY_NAME=ovirt-volume-provisioner
FLEX_CONTAINER_NAME=ovirt-flexvolume-driver
PROVISIONER_CONTAINER_NAME=ovirt-volume-provisioner
AUTOMATION_CONTAINER_NAME=ovirt-openshift-extensions-ci
CLOUD_PROVIDER_NAME=ovirt-cloud-provider

REGISTRY=quay.io/rgolangh
VERSION?=$(shell git describe --tags --always --match "v[0-9]*" | awk -F'-' '{print substr($$1,2) }')
RELEASE?=$(shell git describe --tags --always --match "v[0-9]*" | awk -F'-' '{if ($$2 != "") {print $$2 "." $$3} else {print 1}}')
VERSION_RELEASE=$(VERSION)$(if $(RELEASE),-$(RELEASE))

COMMIT=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

COMMON_ENV=CGO_ENABLED=0 GOOS=linux GOARCH=amd64
COMMON_GO_BUILD_FLAGS=-ldflags '-extldflags "-static"'

TARBALL=ovirt-openshift-extensions-$(VERSION_RELEASE).tar.gz

all: clean deps build test container container-push

build-flex:
	$(COMMON_ENV) $(GOBUILD) \
	$(COMMON_GO_BUILD_FLAGS) \
	-o $(PREFIX)/$(FLEX_DRIVER_BINARY_NAME) \
	-v cmd/$(FLEX_DRIVER_BINARY_NAME)/*.go

build-provisioner:
	$(COMMON_ENV) $(GOBUILD) \
	$(COMMON_GO_BUILD_FLAGS) \
	-o $(PREFIX)/$(PROVISIONER_BINARY_NAME) \
	-v cmd/$(PROVISIONER_BINARY_NAME)/*.go

build-cloud-provider:
	$(COMMON_ENV) $(GOBUILD) \
	$(COMMON_GO_BUILD_FLAGS) \
	-o $(PREFIX)/$(CLOUD_PROVIDER_NAME) \
	-v cmd/$(CLOUD_PROVIDER_NAME)/*.go

container: \
	container-flexdriver \
	container-provisioner \
	container-cloud-provider \
	container-ci

container-flexdriver: tarball
	docker build \
		-t $(REGISTRY)/$(FLEX_CONTAINER_NAME):$(VERSION_RELEASE) \
		-t $(REGISTRY)/$(FLEX_CONTAINER_NAME):latest \
		--build-arg VERSION=$(VERSION) \
		--build-arg RELEASE=$(RELEASE) \
		-f deployment/ovirt-flexvolume-driver/container/Dockerfile \
		.

container-provisioner: tarball
	docker build \
		-t $(REGISTRY)/$(PROVISIONER_CONTAINER_NAME):$(VERSION_RELEASE) \
		-t $(REGISTRY)/$(PROVISIONER_CONTAINER_NAME):latest \
		--build-arg VERSION=$(VERSION) \
		--build-arg RELEASE=$(RELEASE) \
		-f deployment/ovirt-volume-provisioner/container/Dockerfile \
		.

container-cloud-provider: tarball
	docker build \
		-t $(REGISTRY)/$(CLOUD_PROVIDER_NAME):$(VERSION_RELEASE) \
		-t $(REGISTRY)/$(CLOUD_PROVIDER_NAME):latest \
		--build-arg VERSION=$(VERSION) \
		--build-arg RELEASE=$(RELEASE) \
		-f deployment/$(CLOUD_PROVIDER_NAME)/container/Dockerfile \
		.

container-ci:
	docker build \
		-t $(REGISTRY)/$(AUTOMATION_CONTAINER_NAME):$(VERSION_RELEASE) \
		-t $(REGISTRY)/$(AUTOMATION_CONTAINER_NAME):latest \
		-f automation/ci/Dockerfile \
		automation/ci

container-push:
	@docker login -u rgolangh -p ${QUAY_API_KEY} quay.io
	docker push $(REGISTRY)/$(FLEX_CONTAINER_NAME):$(VERSION_RELEASE)
	docker push $(REGISTRY)/$(PROVISIONER_CONTAINER_NAME):$(VERSION_RELEASE)
	docker push $(REGISTRY)/$(AUTOMATION_CONTAINER_NAME):$(VERSION_RELEASE)
	docker push $(REGISTRY)/$(CLOUD_PROVIDER_NAME):$(VERSION_RELEASE)
	# push latest
	docker push $(REGISTRY)/$(FLEX_CONTAINER_NAME):latest
	docker push $(REGISTRY)/$(PROVISIONER_CONTAINER_NAME):latest
	docker push $(REGISTRY)/$(AUTOMATION_CONTAINER_NAME):latest
	docker push $(REGISTRY)/$(CLOUD_PROVIDER_NAME):latest

apb_build:
	$(MAKE) -C deployment/ovirt-flexvolume-driver-apb/ apb_build REGISTRY=$(REGISTRY)

apb_push:
	$(MAKE) -C deployment/ovirt-flexvolume-driver-apb/ apb_push REGISTRY=$(REGISTRY)

apb_docker_push:
	$(MAKE) -C deployment/ovirt-flexvolume-driver-apb/ docker_push REGISTRY=$(REGISTRY)

build: \
	build-flex \
	build-provisioner \
	build-cloud-provider

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	git clean -dfx -e .idea*

deps:
	dep ensure --update

tarball:
	/bin/git archive --format=tar.gz HEAD > $(TARBALL)

.PHONY: build-flex build-provisioner build-cloud-provider container container-flexdriver container-provisioner container-cloud-provider container-ci container-push
