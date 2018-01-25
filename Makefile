# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GODEP=dep

BINARY_NAME=ovirt-flexdriver

VERSION?=?
COMMIT=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

all: clean deps build test
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -a -ldflags '-extldflags "-static"' -o $(BINARY_NAME) -v cmd/$(BINARY_NAME)/*.go
test:
	$(GOTEST) -v ./...
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
run:
	build
	./$(BINARY_NAME)
deps:
	$(GOGET) github.com/golang/dep/cmd/dep
	$(GODEP) ensure

# Cross compilation
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -a -ldflags '-extldflags "-static"' -o $(BINARY_NAME) -v
