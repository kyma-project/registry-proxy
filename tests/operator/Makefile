PROJECT_ROOT = ../..
include ${PROJECT_ROOT}/hack/help.mk

.PHONY: test
test: ## Run integration test.
	go run main.go

#TODO: refactor, common part with operator
.PHONY: cluster-info
cluster-info: ## Print useful info about the cluster regarding integration run
	@echo "####################### Operator Logs #######################"
	@kubectl logs -n kyma-system -l app.kubernetes.io/name=registry-proxy-operator --tail=-1 || true
	@echo ""

	@echo "####################### Registry Proxy Logs #######################"
	@kubectl logs -n kyma-system -l app.kubernetes.io/name=registry-proxy --tail=-1 || true
	@echo ""

	@echo "####################### Registry Proxies CR #######################"
	@kubectl get registryproxies.operator.kyma-project.io -A -oyaml || true
	@echo ""

	@echo "####################### Connection CR #######################"
	@kubectl get connections.registry-proxy.kyma-project.io -A -oyaml || true
	@echo ""

	@echo "####################### Pods #######################"
	@kubectl get pods -A || true
	@echo ""
