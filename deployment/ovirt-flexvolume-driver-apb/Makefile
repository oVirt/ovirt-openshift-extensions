# set the registry from the main Makefile under the root of the repo
REGISTRY= 
IMAGENAME = ovirt-flexvolume-driver-apb
TAG = latest
USER=$(shell id -u)
PWD=$(shell pwd)
KUBEDIR=$(HOME)/.kube

build_and_push: apb_build docker_push apb_push

.PHONY: apb_build
apb_build:
	docker run --rm --privileged -v $(PWD):/mnt:z -v /var/run/docker.sock:/var/run/docker.sock -u ${USER} docker.io/ansibleplaybookbundle/apb-tools:latest prepare
	docker build -t $(REGISTRY)/$(IMAGENAME):$(TAG) .

.PHONY: docker_push
docker_push:
	docker push $(REGISTRY)/$(IMAGENAME):$(TAG)

.PHONY: apb_push
apb_push:
	docker run --rm --privileged -v $(PWD):/mnt:z -v $(KUBEDIR):/.kube:z -v /var/run/docker.sock:/var/run/docker.sock -u 1000 docker.io/ansibleplaybookbundle/apb-tools:latest push
