package state

import (
	"context"
	"fmt"

	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/fsm"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func sFnConnectivityProxyURL(ctx context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	if m.State.Connection.Spec.ProxyURL != "" {
		m.State.ProxyURL = m.State.Connection.Spec.ProxyURL
	} else {
		proxyURL, err := getReverseProxyURL(ctx, m)
		if err != nil {
			return stopWithEventualError(err)
		}
		m.State.ProxyURL = proxyURL
	}
	return nextState(sFnHandleDeployment)
}

func getReverseProxyURL(ctx context.Context, m *fsm.StateMachine) (string, error) {
	connectivityProxyKey := client.ObjectKey{
		Name:      "connectivity-proxy",
		Namespace: "kyma-system",
	}

	connectivityProxy := &unstructured.Unstructured{}
	connectivityProxy.SetNamespace(connectivityProxyKey.Namespace)
	connectivityProxy.SetName(connectivityProxyKey.Name)
	connectivityProxy.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "connectivityproxy.sap.com",
		Version: "v1",
		Kind:    "ConnectivityProxy",
	})

	err := m.Client.Get(ctx, connectivityProxyKey, connectivityProxy)
	if err != nil {
		return "", err
	}

	proxyPort, found, err := unstructured.NestedFieldCopy(connectivityProxy.Object, "spec", "config", "servers", "proxy", "http", "port")
	if err != nil {
		return "", fmt.Errorf("failed to get proxy port from connectivity proxy: %v", err)
	}
	if !found {
		return "", fmt.Errorf("proxy http port was not specified in the connectivity proxy")
	}
	proxyURL := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", connectivityProxyKey.Name, connectivityProxyKey.Namespace, proxyPort.(int64))
	return proxyURL, nil
}
