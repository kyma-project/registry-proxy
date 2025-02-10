## Location to install dependencies to
ifndef PROJECT_ROOT
$(error PROJECT_ROOT is undefined)
endif
LOCALBIN ?= $(realpath $(PROJECT_ROOT))/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUBECTL ?= kubectl
KUSTOMIZE ?= $(LOCALBIN)/kustomize
KUBEBUILDER ?= $(LOCALBIN)/kubebuilder
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint

## Tool Versions
KUSTOMIZE_VERSION ?= v5.6.0
KUBEBUILDER_VERSION ?= v4.5.0
CONTROLLER_TOOLS_VERSION ?= v0.16.4
ENVTEST_VERSION ?= release-0.19
GOLANGCI_LINT_VERSION ?= v1.61.0

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v5,$(KUSTOMIZE_VERSION))

.PHONY: kubebuilder
kubebuilder: $(KUBEBUILDER) ## Download controller-gen locally if necessary.
$(KUBEBUILDER): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/kubebuilder/v4/cmd@$(KUBEBUILDER_VERSION)
	mv $(LOCALBIN)/cmd $(KUBEBUILDER)
	ln -sf $(KUBEBUILDER) $(KUBEBUILDER)-$(KUBEBUILDER_VERSION)

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))

.PHONY: envtest
envtest: $(ENVTEST) ## Download setup-envtest locally if necessary.
$(ENVTEST): $(LOCALBIN)
	$(call go-install-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest,$(ENVTEST_VERSION))

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f $(1) || true ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv $(1) $(1)-$(3) ;\
} ;\
ln -sf $(1)-$(3) $(1)
endef


########## Envtest ###########
ENVTEST ?= $(LOCALBIN)/setup-envtest
KUBEBUILDER_ASSETS=$(LOCALBIN)/k8s/kubebuilder_assets

define path_error
$(error Error: path is empty: $1)
endef
