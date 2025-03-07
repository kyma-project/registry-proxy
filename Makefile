# Image URL to use all building/pushing image targets
CTRL_IMG ?= image-pull-reverse-proxy-controller:main
IMG ?= image-pull-reverse-proxy:main
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

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix

.PHONY: generate
generate: ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations for controller.
	make -C components/controller generate

.PHONY: manifests
manifests: controller-gen kubebuilder ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	rm -rf config || true
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./components/controller/..." output:crd:artifacts:config=config/crd/bases
	rm config/crd/bases/operator.kyma-project.io_connectivityproxies.yaml
	$(KUBEBUILDER) edit --plugins=helm/v1-alpha

.PHONY: run-local
run-local: create-k3d ## Setup local k3d cluster and install controller
	make -C components/controller docker-build CTRL_IMG=$(CTRL_IMG)
	k3d image import $(CTRL_IMG) -c kyma
	make -C components/reverse-proxy docker-build IMG=$(IMG)
	k3d image import $(IMG) -c kyma
	## make sure helm is installed or binary is present
	helm install image-pull-reverse-proxy-controller $(PROJECT_ROOT)/dist/chart

.PHONY: apply-default-cr
apply-default-cr: manifests ## Apply default CustomResource
	kubectl apply -f examples/example-cr.yaml

.PHONY: test
test: ## Run unit tests
	make -C components/controller test
	make -C components/reverse-proxy test
