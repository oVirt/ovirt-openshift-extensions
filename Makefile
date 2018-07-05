# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GODEP=dep

PREFIX=.
ARTIFACT_DIR ?= .

FLEX_DRIVER_BINARY_NAME=ovirt-flexdriver
PROVISIONER_BINARY_NAME=ovirt-provisioner
FLEX_CONTAINER_NAME=ovirt-flexvolume-driver
PROVISIONER_CONTAINER_NAME=ovirt-volume-provisioner

REGISTRY=rgolangh
VERSION?=$(shell git describe --tags --always --match "v[0-9]*" | awk -F'-' '{print $$1 }')
RELEASE?=$(shell git describe --tags --always --match "v[0-9]*" | awk -F'-' '{print $$2 "." $$3}')
COMMIT=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

COMMON_ENV=CGO_ENABLED=0 GOOS=linux GOARCH=amd64
COMMON_GO_BUILD_FLAGS=-a -ldflags '-extldflags "-static"'

TARBALL=ovirt-openshift-extensions-$(VERSION)$(if $(RELEASE),-$(RELEASE)).tar.gz

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

container: \
	container-flexdriver \
	container-provisioner

container-flexdriver:
	# place the rpm flat under the repo otherwise dockerignore will mask its directory. TODO make it more flexible
	docker build \
         -t $(REGISTRY)/$(FLEX_CONTAINER_NAME):$(VERSION)$(if $(RELEASE),_$(RELEASE)) \
         -t $(REGISTRY)/$(FLEX_CONTAINER_NAME):latest \
         . \
         -f deployment/ovirt-flexdriver/container/Dockerfile

container-provisioner:
	docker build \
        -t $(REGISTRY)/$(PROVISIONER_CONTAINER_NAME):$(VERSION)$(if $(RELEASE),_$(RELEASE)) \
        -t $(REGISTRY)/$(PROVISIONER_CONTAINER_NAME):latest \
        . \
        -f deployment/ovirt-provisioner/container/Dockerfile

container-push:
	@docker login -u rgolangh -p ${DOCKER_BUILDER_API_KEY}
	docker push $(REGISTRY)/$(FLEX_CONTAINER_NAME):$(VERSION)$(if $(RELEASE),_$(RELEASE))
	docker push $(REGISTRY)/$(PROVISIONER_CONTAINER_NAME):$(VERSION)$(if $(RELEASE),_$(RELEASE))
	# push latest
	docker push $(REGISTRY)/$(FLEX_CONTAINER_NAME):latest
	docker push $(REGISTRY)/$(PROVISIONER_CONTAINER_NAME):latest

apb_build:
	$(MAKE) -C deployment/ovirt-flexvolume-driver-apb/ apb_build

apb_push:
	$(MAKE) -C deployment/ovirt-flexvolume-driver-apb/ apb_push

apb_docker_push:
	$(MAKE) -C deployment/ovirt-flexvolume-driver-apb/ docker_push

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
	dep ensure

tarball:
	/bin/git archive --format=tar.gz HEAD > $(TARBALL)

rpm:
	$(MAKE) tarball
	rpmbuild -tb $(TARBALL) \
		--define "debug_package %{nil}" \
		--define "_rpmdir $(ARTIFACT_DIR)" \
		--define "_version $(VERSION)" \
		--define "_release $(RELEASE)"

srpm:
	$(MAKE) tarball
	rpmbuild -ts $(TARBALL) \
		--define "debug_package %{nil}" \
		--define "_rpmdir $(ARTIFACT_DIR)" \
		--define "_version $(VERSION)" \
		--define "_release $(RELEASE)"

.PHONY: build-flex build-provisioner container container-flexdriver container-provisioner container-provisioner-binary container-provisioner-ansible container-push
