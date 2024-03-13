# Copyright (c) NVIDIA CORPORATION.  All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

DOCKER ?= docker

include $(CURDIR)/versions.mk

DOCKERFILE_DEVEL := "images/devel/Dockerfile"
K8S_TEST_INFRA := "https://github.com/NVIDIA/k8s-test-infra.git"

TARGETS := binary build all check fmt assert-fmt generate test coverage golangci-lint
DOCKER_TARGETS := $(patsubst %, docker-%, $(TARGETS))
.PHONY: $(TARGETS) $(DOCKER_TARGETS)

GOOS := linux

build:
	GOOS=$(GOOS) go build ./...

all: check build binary
check: golangci-lint

# Apply go fmt to the codebase
fmt:
	go list -f '{{.Dir}}' $(MODULE)/... \
		| xargs gofmt -s -l -w

assert-fmt:
	go list -f '{{.Dir}}' $(MODULE)/... \
		| xargs gofmt -s -l > fmt.out
	@if [ -s fmt.out ]; then \
		echo "\nERROR: The following files are not formatted:\n"; \
		cat fmt.out; \
		rm fmt.out; \
		exit 1; \
	else \
		rm fmt.out; \
	fi

generate:
	go generate $(MODULE)/...

golangci-lint:
	golangci-lint run ./...

COVERAGE_FILE := coverage.out
test: build
	go test -v -coverprofile=$(COVERAGE_FILE) $(MODULE)/...

coverage: test
	cat $(COVERAGE_FILE) | grep -v "_mock.go" > $(COVERAGE_FILE).no-mocks
	go tool cover -func=$(COVERAGE_FILE).no-mocks

build-image: $(DOCKERFILE_DEVEL)
	$(DOCKER) build \
		--progress=plain \
		--build-arg GOLANG_VERSION="$(GOLANG_VERSION)" \
		--build-arg CLIENT_GEN_VERSION="$(CLIENT_GEN_VERSION)" \
		--build-arg CONTROLLER_GEN_VERSION="$(CONTROLLER_GEN_VERSION)" \
		--build-arg GOLANGCI_LINT_VERSION="$(GOLANGCI_LINT_VERSION)" \
		--build-arg MOQ_VERSION="$(MOQ_VERSION)" \
		--tag $(BUILDIMAGE) \
		-f $(DOCKERFILE_DEVEL) \
		.

$(DOCKER_TARGETS): docker-%:
	@echo "Running 'make $(*)' in container image $(BUILDIMAGE)"
	$(DOCKER) run \
		--rm \
		-e GOCACHE=/tmp/.cache/go \
		-e GOMODCACHE=/tmp/.cache/gomod \
		-v $(PWD):/work \
		-w /work \
		--user $$(id -u):$$(id -g) \
		$(BUILDIMAGE) \
			make $(*)

# Start an interactive shell using the development image.
PHONY: .shell
.shell:
	$(DOCKER) run \
		--rm \
		-ti \
		-e GOCACHE=/tmp/.cache/go \
		-e GOMODCACHE=/tmp/.cache/gomod \
		-v $(PWD):/work \
		-w /work \
		--user $$(id -u):$$(id -g) \
		$(BUILDIMAGE)
