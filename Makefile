# Image URL to use all building/pushing image targets
OPERATOR_IMG ?= registry-proxy-operator:main
RP_IMG ?= registry-proxy-controller:main
CONNECTION_IMG ?= registry-proxy-connection:main

IMG_VERSION ?= main
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.31.0

PROJECT_ROOT=.

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

include ${PROJECT_ROOT}/hack/k3d.mk
include ${PROJECT_ROOT}/hack/tools.mk
include ${PROJECT_ROOT}/hack/help.mk

# CONTAINER_TOOL defines the container tool to be used for building images.
# Be aware that the target commands are only tested with Docker which is
# scaffolded by default. However, you might want to replace it to use other
# tools. (i.e. podman)
CONTAINER_TOOL ?= docker

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec


.PHONY: all
all: run-local

##@ Development

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix

.PHONY: generate
generate: ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations for controller.
	make -C components/operator generate
	make -C components/registry-proxy generate

.PHONY: manifests
# TODO: autogenerate in correct place, or mv files
manifests: controller-gen kubebuilder ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	make -C components/operator manifests
	make -C components/registry-proxy manifests

.PHONY: run-local-operator
run-local-operator: create-k3d ## Setup local k3d cluster and install operator
	IMG_DIRECTORY="" IMG_VERSION="${IMG_VERSION}" OPERATOR_IMG=$(OPERATOR_IMG) make -C components/operator docker-build-local
	k3d image import $(OPERATOR_IMG) -c kyma
	RP_IMG=$(RP_IMG) make -C components/registry-proxy docker-build
	k3d image import $(RP_IMG) -c kyma
	CONNECTION_IMG=$(CONNECTION_IMG) make -C components/connection docker-build
	k3d image import $(CONNECTION_IMG) -c kyma
	## make sure helm is installed or binary is present
	helm install registry-proxy-operator $(PROJECT_ROOT)/config/operator --namespace=kyma-system --set controllerManager.container.image="$(OPERATOR_IMG)"
	# TODO: wait for ready status

.PHONY: run-local
run-local: ## Setup local k3d cluster and install operator with sample CR
	kubectl apply -f $(PROJECT_ROOT)/config/samples/default-registry-proxy-cr.yaml

.PHONY: integration-dependencies
integration-dependencies:  ## create k3d cluster and run integration test
	# connectivity proxy
	kubectl apply -f $(PROJECT_ROOT)/hack/connectivity-proxy/stateful-set-mock.yaml
	# docker registry
	kubectl apply -f https://github.com/kyma-project/docker-registry/releases/latest/download/dockerregistry-operator.yaml
	kubectl apply -f https://github.com/kyma-project/docker-registry/releases/latest/download/default-dockerregistry-cr.yaml
	# upload test image to the docker registry
	docker pull alpine:3.21.3
	kubectl wait --for condition=Available -n kyma-system deployment dockerregistry-operator --timeout=90s
	sleep 5
	kubectl wait --for condition=Available -n kyma-system deployment dockerregistry --timeout=60s
	sleep 5
	$(KYMA) alpha registry image-import alpine:3.21.3

.PHONY: run-integration-test-operator
run-integration-test-operator: integration-dependencies ## run integration test
	make -C tests/operator test

.PHONY: run-integration-test-registry-proxy
run-integration-test-registry-proxy: integration-dependencies 
	make -C tests/registry-proxy test

# TODO: move somewhere else
.PHONY: install-registry-proxy
install-registry-proxy:
	helm install registry-proxy-controller $(PROJECT_ROOT)/config/registry-proxy \
	--namespace=kyma-system \
	# --set controllerManager.container.image="europe-docker.pkg.dev/kyma-project/prod/registry-proxy-controller:$(TAG)" \
	# --set controllerManager.container.env.PROXY_IMAGE=$(IMG)

.PHONY: integration-tests
integration-tests: ## Run integration tests
	make -C tests/registry-proxy test
	make -C tests/operator test

.PHONY: unit-tests
unit-tests: ## Run unit tests
	make -C components/operator test
	make -C components/registry-proxy test
	make -C components/connection test

.PHONY: cluster-info
cluster-info:
	make -C tests/registry-proxy cluster-info
##@ Actions
.PHONY: module-config
module-config:
	yq ".version = \"${MODULE_VERSION}\""\
    module-config-template.yaml > module-config.yaml
