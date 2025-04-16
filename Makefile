# Image URL to use all building/pushing image targets
CTRL_IMG ?= registry-proxy-controller:main
IMG ?= registry-proxy:main
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
	helm install registry-proxy-controller $(PROJECT_ROOT)/dist/chart


.PHONY: run-local-integration
run-local-integration: run-local
	make -C components/controller docker-build CTRL_IMG=$(CTRL_IMG)
	k3d image import $(CTRL_IMG) -c kyma
	make -C components/reverse-proxy docker-build IMG=$(IMG)
	k3d image import $(IMG) -c kyma
	## make sure helm is installed or binary is present
	helm install registry-proxy-controller $(PROJECT_ROOT)/dist/chart --namespace=kyma-system
	# connectivity proxy
	kubectl apply -f $(PROJECT_ROOT)/hack/connectivity-proxy/connectivity-proxy.yaml
	kubectl apply -f $(PROJECT_ROOT)/hack/connectivity-proxy/connectivity-proxy-default-cr.yaml
	# docker registry
	kubectl apply -f https://github.com/kyma-project/docker-registry/releases/latest/download/dockerregistry-operator.yaml
	kubectl apply -f https://github.com/kyma-project/docker-registry/releases/latest/download/default-dockerregistry-cr.yaml
	# upload test image to the docker registry
	docker pull alpine:3.21.3
	kubectl wait --for condition=Available -n kyma-system deployment dockerregistry-operator --timeout=60s
	sleep 5
	kubectl wait --for condition=Available -n kyma-system deployment dockerregistry --timeout=60s
	sleep 5
	$(KYMA) alpha registry image-import alpine:3.21.3
	go run $(PROJECT_ROOT)/tests/main.go

.PHONY: run-integration
run-integration:
	## make sure helm is installed or binary is present
	helm install registry-proxy-controller $(PROJECT_ROOT)/dist/chart --namespace=kyma-system --set controllerManager.container.image.repository="europe-docker.pkg.dev/kyma-project/prod/registry-proxy-controller" --set controllerManager.container.image.tag=$(TAG) --set controllerManager.container.env.PROXY_IMAGE=$(IMG)
	# connectivity proxy
	kubectl apply -f $(PROJECT_ROOT)/hack/connectivity-proxy/connectivity-proxy.yaml
	kubectl apply -f $(PROJECT_ROOT)/hack/connectivity-proxy/connectivity-proxy-default-cr.yaml
	# docker registry
	kubectl apply -f https://github.com/kyma-project/docker-registry/releases/latest/download/dockerregistry-operator.yaml
	kubectl apply -f https://github.com/kyma-project/docker-registry/releases/latest/download/default-dockerregistry-cr.yaml
	# upload test image to the docker registry
	docker pull alpine:3.21.3
	kubectl wait --for condition=Available -n kyma-system deployment dockerregistry-operator --timeout=60s
	sleep 5
	kubectl wait --for condition=Available -n kyma-system deployment dockerregistry --timeout=60s
	sleep 5
	$(KYMA) alpha registry image-import alpine:3.21.3
	go run $(PROJECT_ROOT)/tests/main.go


.PHONY: apply-default-cr
apply-default-cr: manifests kyma ## Apply default CustomResource
	kubectl apply -f examples/example-cr.yaml
	# load image into the docker registry

.PHONY: test
test: ## Run unit tests
	make -C components/controller test
	make -C components/reverse-proxy test

.PHONY: cluster-info
cluster-info: ## Print useful info about the cluster regarding integration run
	@echo "####################### Controller Logs #######################"
	@kubectl logs -n kyma-system -l app.kubernetes.io/name=registry-proxy --tail=-1 || true
	@echo ""

	@echo "####################### RP CR #######################"
	@kubectl get imagepullreverseproxies -A -oyaml || true
	@echo ""

	@echo "####################### Pods #######################"
	@kubectl get pods -A || true
	@echo ""

##@ Actions
.PHONY: module-config
module-config:
	yq ".channel = \"${CHANNEL}\" | .version = \"${MODULE_VERSION}\""\
    	module-config-template.yaml > module-config.yaml
