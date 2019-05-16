# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GODEP=dep

PREFIX=.
ARTIFACT_DIR ?= .

REGISTRY=quay.io/rgolangh
VERSION?=$(shell git describe --tags --always --match "v[0-9]*" | awk -F'-' '{print substr($$1,2) }')
RELEASE?=$(shell git describe --tags --always --match "v[0-9]*" | awk -F'-' '{if ($$2 != "") {print $$2 "." $$3} else {print 1}}')
VERSION_RELEASE=$(VERSION)$(if $(RELEASE),-$(RELEASE))

COMMIT=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

COMMON_ENV=CGO_ENABLED=0 GOOS=linux GOARCH=amd64
COMMON_GO_BUILD_FLAGS=-ldflags '-extldflags "-static"'

TARBALL=ovirt-openshift-extensions-$(VERSION_RELEASE).tar.gz
PUSH_LATEST=1

all: clean deps build test container container-push

binaries = \
	ovirt-flexvolume-driver \
	ovirt-volume-provisioner \
	ovirt-cloud-provider

containers = \
	$(binaries) \
	ovirt-openshift-installer

$(binaries): test internal
	go vet ./cmd/$@ && \
	$(COMMON_ENV) $(GOBUILD) \
    	$(COMMON_GO_BUILD_FLAGS) \
    	-o $(PREFIX)/$@ \
    	-v cmd/$@/*.go

.PHONY: internal
internal:
	go vet ./internal


container-%: DIR=.
container-%: DOCKERFILE=deployment/$*/container/Dockerfile
container-ovirt-openshift-installer: DIR=automation/ci
container-ovirt-openshift-installer: DOCKERFILE=$(DIR)/Dockerfile

container-%: tarball
	docker build \
		-t $(REGISTRY)/$*:$(VERSION_RELEASE) \
		-t $(REGISTRY)/$*:latest \
		--build-arg version=$(VERSION) \
		--build-arg release=$(RELEASE) \
		-f $(DOCKERFILE) \
		$(DIR)

container-push-%:
	@docker login -u rgolangh+ovirtci -p ${QUAY_API_KEY} quay.io
	docker push $(REGISTRY)/$*:$(VERSION_RELEASE)
ifneq ($(PUSH_LATEST),0)
	docker push $(REGISTRY)/$*:latest
endif
	echo "$(REGISTRY)/$*:$(VERSION_RELEASE)" >> containers-artifacts.list

build: $(binaries)

build-containers: $(addprefix container-, $(containers))

push-containers: $(addprefix container-push-, $(containers))

test:
	$(GOTEST) -v ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o /tmp/coverage.html

clean:
	$(GOCLEAN)
	git clean -dfx -e .idea*

deps:
	dep ensure --update

tarball: $(TARBALL)

$(TARBALL):
	/bin/git archive --format=tar.gz HEAD > $(TARBALL)

apb_build:
	$(MAKE) -C deployment/ovirt-flexvolume-driver-apb/ \
		REGISTRY=$(REGISTRY) \
		VERSION=$(VERSION) \
		RELEASE=$(RELEASE) \
		VERSION_RELEASE=$(VERSION_RELEASE) \
		apb_build

apb_push:
	$(MAKE) -C deployment/ovirt-flexvolume-driver-apb/ apb_push REGISTRY=$(REGISTRY)

apb_docker_push:
	$(MAKE) -C deployment/ovirt-flexvolume-driver-apb/ \
		REGISTRY=$(REGISTRY) \
		VERSION=$(VERSION) \
		RELEASE=$(RELEASE) \
		VERSION_RELEASE=$(VERSION_RELEASE) \
		docker_push

.PHONY: all tarball test build build-containers push-containers apb_build apb_docker_push apb_push
