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
REGISTRY=rgolangh
VERSION?=$(shell git describe --tags --always|cut -d "-" -f1)
RELEASE?=$(shell git describe --tags --always|cut -d "-" -f2- | sed 's/-/_/')
COMMIT=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

COMMON_ENV=CGO_ENABLED=0 GOOS=linux GOARCH=amd64
COMMON_GO_BUILD_FLAGS=-a -ldflags '-extldflags "-static"'

TARBALL=${FLEX_DRIVER_BINARY_NAME}-${VERSION}-${RELEASE}.tar.gz

all: clean deps build test container container-push

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

container: \
	container-flexdriver \
	container-provisioner

container-flexdriver:
	docker build -t $(REGISTRY)/$(FLEX_DRIVER_BINARY_NAME)-ansible:$(VERSION) . -f deployment/ovirt-flexdriver/container/Dockerfile
	docker tag $(REGISTRY)/$(FLEX_DRIVER_BINARY_NAME)-ansible:$(VERSION) $(REGISTRY)/$(FLEX_DRIVER_BINARY_NAME)-ansible:latest

container-provisioner: \
	container-provisioner-binary \
	container-provisioner-ansible

container-provisioner-binary:
	docker build -t $(REGISTRY)/$(PROVISIONER_BINARY_NAME):$(VERSION) . -f deployment/ovirt-provisioner/container/binary/Dockerfile
	docker tag $(REGISTRY)/$(PROVISIONER_BINARY_NAME):$(VERSION) $(REGISTRY)/$(PROVISIONER_BINARY_NAME):latest

container-provisioner-ansible:
	docker build -t $(REGISTRY)/$(PROVISIONER_BINARY_NAME)-ansible:$(VERSION) . -f  deployment/ovirt-provisioner/container/ansible/Dockerfile
	docker tag $(REGISTRY)/$(PROVISIONER_BINARY_NAME)-ansible:$(VERSION) $(REGISTRY)/$(PROVISIONER_BINARY_NAME)-ansible:latest

container-push:
	docker login -u rgolangh -p ${DOCKER_BUILDER_API_KEY}
	docker push $(REGISTRY)/$(FLEX_DRIVER_BINARY_NAME)-ansible:$(VERSION)
	docker push $(REGISTRY)/$(PROVISIONER_BINARY_NAME):$(VERSION)
	docker push $(REGISTRY)/$(PROVISIONER_BINARY_NAME)-ansible:$(VERSION)
	# push latest
	docker push $(REGISTRY)/$(FLEX_DRIVER_BINARY_NAME)-ansible:latest
	docker push $(REGISTRY)/$(PROVISIONER_BINARY_NAME):latest
	docker push $(REGISTRY)/$(PROVISIONER_BINARY_NAME)-ansible:latest

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
	glide --debug  install --strip-vendor

rpm:
	/bin/git ls-files | tar --files-from /proc/self/fd/0 -czf "$(TARBALL)"
ifdef ARTIFACT_DIR
	rpmbuild -tb $(TARBALL) \
	    --define "debug_package %{nil}" \
        --define "_rpmdir ${ARTIFACT_DIR}" \
	    --define "_version ${VERSION}" \
	    --define "_release ${RELEASE}"
else
	rpmbuild -tb $(TARBALL) \
			--define "debug_package %{nil}" \
			--define "_version ${VERSION}" \
			--define "_release ${RELEASE}"
endif

.PHONY: build-flex build-provisioner container container-flexdriver container-provisioner container-provisioner-binary container-provisioner-ansible container-push
