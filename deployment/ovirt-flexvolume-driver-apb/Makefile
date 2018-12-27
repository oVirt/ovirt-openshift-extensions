# set the registry from the main Makefile under the root of the repo
REGISTRY=
VERSION=
RELEASE=
VERSION_RELEASE=
IMAGENAME = ovirt-flexvolume-driver-apb
USER=$(shell id -u)
PWD=$(shell pwd)
KUBEDIR=$(HOME)/.kube

build_and_push: apb_build docker_push apb_push

.PHONY: apb_build
apb_build:
	docker run \
		--rm --privileged \
		-v $(PWD):/mnt:z -v /var/run/docker.sock:/var/run/docker.sock \
		-u ${USER} \
		docker.io/ansibleplaybookbundle/apb-tools:latest prepare
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg RELEASE=$(RELEASE) \
		-t $(REGISTRY)/$(IMAGENAME):latest \
		-t $(REGISTRY)/$(IMAGENAME):$(VERSION_RELEASE) \
		.

.PHONY: docker_push
docker_push:
	docker push $(REGISTRY)/$(IMAGENAME):latest
	docker push $(REGISTRY)/$(IMAGENAME):$(VERSION_RELEASE)

.PHONY: apb_push
apb_push:
	docker run \
		--rm --privileged \
		-v $(PWD):/mnt:z \
		-v $(KUBEDIR):/.kube:z \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-u 1000 docker.io/ansibleplaybookbundle/apb-tools:latest push
