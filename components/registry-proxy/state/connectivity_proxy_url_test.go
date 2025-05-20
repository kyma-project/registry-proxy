package state

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/api/v1alpha1"
	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/fsm"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func TestGetReverseProxyURL(t *testing.T) {
	t.Run("Connectivity proxy doesn't exist", func(t *testing.T) {
		scheme := minimalScheme(t)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

		m := fsm.StateMachine{
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}
		proxyURL, err := getReverseProxyURL(context.Background(), &m)
		require.NotNil(t, err)
		require.ErrorContains(t, err, "\"connectivity-proxy\" not found")
		require.Equal(t, "", proxyURL)
	})

	t.Run("Conenctivity proxy exists", func(t *testing.T) {
		scheme := minimalScheme(t)

		connectivityProxy := minimalConnectivityProxy(8080)

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(connectivityProxy).Build()
		m := fsm.StateMachine{
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}

		proxyURL, err := getReverseProxyURL(context.Background(), &m)
		require.Nil(t, err)
		require.Equal(t, "http://connectivity-proxy.kyma-system.svc.cluster.local:8080", proxyURL)
	})

	t.Run("Conenctivity proxy exists, but http proxy is missing", func(t *testing.T) {
		scheme := minimalScheme(t)

		connectivityProxy := minimalConnectivityProxy(0)

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(connectivityProxy).Build()
		m := fsm.StateMachine{
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}

		proxyURL, err := getReverseProxyURL(context.Background(), &m)
		require.NotNil(t, err)
		require.ErrorContains(t, err, "proxy http port was not specified in the connectivity proxy")
		require.Equal(t, "", proxyURL)
	})
}

func Test_sFnConnectivityProxyURL(t *testing.T) {

	t.Run("user provided proxyURL", func(t *testing.T) {
		rp := v1alpha1.Connection{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "rp",
				Namespace: "maslo",
			},
			Spec: v1alpha1.ConnectionSpec{
				ProxyURL:   "http://test-proxy-url",
				TargetHost: "dummy",
			},
		}

		scheme := minimalScheme(t)
		getWasCalled := false
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{
			Get: func(_ context.Context, _ client.WithWatch, _ client.ObjectKey, _ client.Object, _ ...client.GetOption) error {
				getWasCalled = true
				return nil
			},
		}).Build()

		m := fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: rp,
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}
		next, result, err := sFnConnectivityProxyURL(context.Background(), &m)

		require.Nil(t, err)
		require.Nil(t, result)
		require.NotNil(t, next)
		requireEqualFunc(t, sFnHandleDeployment, next)
		require.False(t, getWasCalled)
		require.Equal(t, rp.Spec.ProxyURL, m.State.ProxyURL)
	})

	t.Run("proxyURL from connectivity proxy", func(t *testing.T) {
		rp := v1alpha1.Connection{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "rp",
				Namespace: "maslo",
			},
			Spec: v1alpha1.ConnectionSpec{
				TargetHost: "dummy",
			},
		}
		connectivityProxy := minimalConnectivityProxy(8080)

		scheme := minimalScheme(t)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(connectivityProxy).Build()

		m := fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: rp,
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}
		next, result, err := sFnConnectivityProxyURL(context.Background(), &m)

		require.Nil(t, err)
		require.Nil(t, result)
		require.NotNil(t, next)
		requireEqualFunc(t, sFnHandleDeployment, next)
		require.Equal(t, "http://connectivity-proxy.kyma-system.svc.cluster.local:8080", m.State.ProxyURL)
	})

	t.Run("proxyURL missing", func(t *testing.T) {
		rp := v1alpha1.Connection{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "rp",
				Namespace: "maslo",
			},
			Spec: v1alpha1.ConnectionSpec{
				TargetHost: "dummy",
			},
		}
		connectivityProxy := minimalConnectivityProxy(0)

		scheme := minimalScheme(t)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(connectivityProxy).Build()

		m := fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: rp,
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}
		next, result, err := sFnConnectivityProxyURL(context.Background(), &m)

		require.NotNil(t, err)
		require.ErrorContains(t, err, "proxy http port was not specified in the connectivity proxy")
		require.Nil(t, result)
		require.Nil(t, next)
		require.Equal(t, "", m.State.ProxyURL)
	})
}

func minimalConnectivityProxy(port int64) *unstructured.Unstructured {
	connectivityProxy := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "connectivityproxy.sap.com/v1",
			"kind":       "ConnectivityProxy",
			"metadata": map[string]interface{}{
				"name":      "connectivity-proxy",
				"namespace": "kyma-system",
			},
			"spec": map[string]interface{}{
				"config": map[string]interface{}{
					"servers": map[string]interface{}{
						"proxy": map[string]interface{}{},
					},
				},
			},
		},
	}
	if port != 0 {
		connectivityProxy.Object["spec"].(map[string]interface{})["config"].(map[string]interface{})["servers"].(map[string]interface{})["proxy"].(map[string]interface{})["http"] = map[string]interface{}{
			"port": port,
		}
	}

	return connectivityProxy
}
