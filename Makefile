# If the USE_SUDO_FOR_DOCKER env var is set, prefix docker commands with 'sudo'
ifdef USE_SUDO_FOR_DOCKER
SUDO_CMD = sudo
endif

IMAGE ?= aws-servicebroker:latest
HELM_URL ?= https://awsservicebroker.s3.amazonaws.com/charts
S3URI ?= $(shell echo $(HELM_URL)/ | sed 's/https:/s3:/' | sed 's/.s3.amazonaws.com//')

build: ## Builds the starter pack
	go build -i github.com/awslabs/aws-service-broker/cmd/servicebroker

test: ## Runs the tests
	go test -v $(shell go list ./... | grep -v /vendor/ | grep -v /test/)

functional-test: ## Builds and execs a minikube image for functional testing
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build -o functional-testing/aws-servicebroker --ldflags="-s" github.com/awslabs/aws-service-broker/cmd/servicebroker && \
    cd functional-testing ; \
      docker build -t aws-sb:functest . && \
      docker run --privileged -it --rm aws-sb:functest /start.sh ; \
    cd ../

linux: ## Builds a Linux executable
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
	go build -o servicebroker-linux --ldflags="-s" github.com/awslabs/aws-service-broker/cmd/servicebroker

cf: ## Builds a PCF tile and bosh release
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build -o packaging/cloudfoundry/resources/cfnsb --ldflags="-s" github.com/awslabs/aws-service-broker/cmd/servicebroker && \
	cd packaging/cloudfoundry/ ; \
	  tile build ; \
	cd ../../

image: ## Builds docker image
	docker build . -t $(IMAGE)

clean: ## Cleans up build artifacts
	rm -f servicebroker
	rm -f servicebroker-linux
	rm -f functional-testing/aws-servicebroker
	rm -rf packaging/cloudfoundry/product
	rm -rf packaging/cloudfoundry/release
	rm -f packaging/helm/index.yaml
	rm -f packaging/helm/aws-servicebroker-*.tgz

helm: ## Creates helm release and repository index file
	cd packaging/helm/ ; \
	helm package aws-servicebroker && \
		helm repo index . --url $(HELM_URL) ; \
	cd ../../

deploy-chart: ## Deploys helm chart and index file to S3 path specified by HELM_URL
	make image && \
	docker push $(IMAGE) && \
	make helm && \
	aws s3 cp packaging/helm/aws-servicebroker-*.tgz s3://awsservicebroker/charts/ --acl public-read --profile apbdev && \
	aws s3 cp packaging/helm/index.yaml s3://awsservicebroker/charts/ --acl public-read --profile apbdev

help: ## Shows the help
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
        awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ''

.PHONY: build test functional-test linux cf image helm deploy-chart clean help
