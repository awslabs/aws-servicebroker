# If the USE_SUDO_FOR_DOCKER env var is set, prefix docker commands with 'sudo'
ifdef USE_SUDO_FOR_DOCKER
SUDO_CMD = sudo
endif

IMAGE ?= docker.io/projectheliostest/nextgen-broker
TAG ?= $(shell git describe --tags --always)
PULL ?= IfNotPresent

build: ## Builds the starter pack
	go build -i github.com/jaymccon/cfnsb/cmd/servicebroker

test: ## Runs the tests
	go test -v $(shell go list ./... | grep -v /vendor/ | grep -v /test/)

functional-test: ## Builds and execs a minikube image for functional testing
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build -o functional-testing/aws-servicebroker --ldflags="-s" github.com/jaymccon/cfnsb/cmd/servicebroker && \
    cd functional-testing ; \
      docker build -t aws-sb:functest . && \
      docker run --privileged -it --rm aws-sb:functest /start.sh ; \
    cd ../

linux: ## Builds a Linux executable
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
	go build -o servicebroker-linux --ldflags="-s" github.com/jaymccon/cfnsb/cmd/servicebroker

cf: ## Builds a PCF tile and bosh release
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build -o packaging/cloudfoundry/resources/cfnsb --ldflags="-s" github.com/jaymccon/cfnsb/cmd/servicebroker && \
	cd packaging/cloudfoundry/ ; \
	  tile build ; \
	cd ../../

clean: ## Cleans up build artifacts
	rm -f servicebroker
	rm -f servicebroker-linux
	rm -f functional-testing/aws-servicebroker
	rm -rf packaging/cloudfoundry/product
	rm -rf packaging/cloudfoundry/release

help: ## Shows the help
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
        awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ''

.PHONY: build test functional-test linux cf clean help
