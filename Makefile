# Copyright 2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License"). You
# may not use this file except in compliance with the License. A copy of
# the License is located at
#
# 	http://aws.amazon.com/apache2.0/
#
# or in the "license" file accompanying this file. This file is
# distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF
# ANY KIND, either express or implied. See the License for the specific
# language governing permissions and limitations under the License.

all: docker 

ROOT := $(shell pwd)
BINARY=bin/ecs-secrets
SOURCEDIR=./
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')
LINUX_BINARY=bin/linux-amd64/ecs-secrets
DARWIN_BINARY=bin/darwin-amd64/ecs-secrets

.PHONY: build
build: $(BINARY)

$(BINARY): $(SOURCES)
	./scripts/build

.PHONY: test
test: 
	GO15VENDOREXPERIMENT=1 go test -timeout=120s -v -cover ./modules/... 

.PHONY: generate-deps
generate-deps:
	go get github.com/tools/godep
	go get github.com/golang/mock/mockgen
	go install github.com/golang/mock/mockgen
	go get golang.org/x/tools/cmd/goimports


.PHONY: generate
generate: $(SOURCES)
	GO15VENDOREXPERIMENT=1 PATH=$(PATH):$(ROOT)/scripts go generate ./modules/...

.PHONY: docker-build
docker-build:
	docker run -v $(shell pwd):/usr/src/app/src/github.com/awslabs/ecs-secrets \
		--workdir=/usr/src/app/src/github.com/awslabs/ecs-secrets \
		--env GOPATH=/usr/src/app \
		--env ECS_SECRETS_RELEASE=$(ECS_SECRETS_RELEASE) \
		golang:1.6 make $(LINUX_BINARY)
	docker run -v $(shell pwd):/usr/src/app/src/github.com/awslabs/ecs-secrets \
		--workdir=/usr/src/app/src/github.com/awslabs/ecs-secrets \
		--env GOPATH=/usr/src/app \
		--env ECS_SECRETS_RELEASE=$(ECS_SECRETS_RELEASE) \
		golang:1.6 make $(DARWIN_BINARY)

.PHONY: supported-platforms
supported-platforms: $(LINUX_BINARY) $(DARWIN_BINARY)

$(LINUX_BINARY): $(SOURCES)
	@mkdir -p ./out/linux-amd64
	GO15VENDOREXPERIMENT=1 TARGET_GOOS=linux GOARCH=amd64 ./scripts/build true ./out/linux-amd64
	@echo "Built ecs-secrets for linux"

$(DARWIN_BINARY): $(SOURCES)
	@mkdir -p ./out/darwin-amd64
	GO15VENDOREXPERIMENT=1 TARGET_GOOS=darwin GOARCH=amd64 ./scripts/build true ./out/darwin-amd64
	@echo "Built ecs-secrets for darwin"

.PHONY: build-in-docker 
build-in-docker:
	@docker build -f scripts/dockerfiles/Dockerfile.build -t "amazon/amazon-ecs-secrets-build:make" .
	@docker run --net=none -v "$(shell pwd)/out:/out" -v "$(shell pwd):/go/src/github.com/awslabs/ecs-secrets" "amazon/amazon-ecs-secrets-build:make"

.PHONY: docker 
docker: certs build-in-docker
	@cd scripts && ./create-ecs-secrets-scratch
	@docker build -f scripts/dockerfiles/Dockerfile.release -t "amazon/amazon-ecs-secrets:make" .
	@echo "Built Docker image \"amazon/amazon-ecs-secrets:make\""

.PHONY: certs
certs: misc/certs/ca-certificates.crt
misc/certs/ca-certificates.crt:
	docker build -t "amazon/amazon-ecs-secrets-cert-source:make" misc/certs/
	docker run "amazon/amazon-ecs-secrets-cert-source:make" cat /etc/ssl/certs/ca-certificates.crt > misc/certs/ca-certificates.crt

.PHONY: docker-release 
docker-release:
	@docker build -f scripts/dockerfiles/Dockerfile.cleanbuild -t "amazon/amazon-ecs-secrets-cleanbuild:make" .
	@echo "Built Docker image \"amazon/amazon-ecs-secrets-cleanbuild:make\""
	@docker run --net=none -v "$(shell pwd)/out:/out" -v "$(shell pwd):/src/ecs-secrets" "amazon/amazon-ecs-secrets-cleanbuild:make"

.PHONY: release 
release: certs docker-release
	@cd scripts && ./create-ecs-secrets-scratch
	@docker build -f scripts/dockerfiles/Dockerfile.release -t "amazon/amazon-ecs-secrets:latest" .
	@echo "Built Docker image \"amazon/amazon-ecs-secrets:latest\""
.PHONY: clean
clean:
	rm -f misc/certs/ca-certificates.crt &> /dev/null
	rm -rf ./out/ ||:
