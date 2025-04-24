package state

import (
	"context"
	"testing"
	"time"

	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/api/v1alpha1"
	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/cache"
	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/fsm"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_sFnValidateReverseProxyURL(t *testing.T) {
	t.Run("when function is valid should go to the next state", func(t *testing.T) {
		m := fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: v1alpha1.RegistryProxy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "rp",
						Namespace: "maslo",
					},
					Spec: v1alpha1.RegistryProxySpec{
						ProxyURL: "http://test-proxy-url",
					},
				},
			},
		}

		next, result, err := sFnValidateReverseProxyURL(context.Background(), &m)

		// Assert
		// no errors
		require.Nil(t, err)
		// without stopping processing
		require.Nil(t, result)
		// with expected next state
		require.NotNil(t, next)
		requireEqualFunc(t, sFnHandleDeployment, next)
		// function conditions remain unchanged
		require.Empty(t, m.State.RegistryProxy.Status.Conditions)
	})
	t.Run("when function is invalid should stop processing", func(t *testing.T) {
		m := fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: v1alpha1.RegistryProxy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "rp",
						Namespace: "maslo",
					},
					Spec: v1alpha1.RegistryProxySpec{
						ProxyURL: ":thisURLisbroken",
					},
				},
			},
		}

		next, result, err := sFnValidateReverseProxyURL(context.Background(), &m)

		// Assert
		// no errors
		require.Nil(t, err)
		// without stopping processing
		require.Nil(t, result)
		// with expected next state
		require.Nil(t, next)
		// function has proper condition
		requireContainsCondition(t, m.State.RegistryProxy.Status,
			v1alpha1.ConditionReady,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonInvalidProxyURL,
			"Invalid Connectivity Proxy URL: parse \":thisURLisbroken\": missing protocol scheme")

	})
}

func Test_sFnValidateConnectivityProxyCRD(t *testing.T) {
	t.Run("when Connectivity Proxy CRD is not installed should update condition and requeue", func(t *testing.T) {
		m := fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: v1alpha1.RegistryProxy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "rp",
						Namespace: "maslo",
					},
				},
			},
			Cache: cache.NewInMemoryBoolCache(),
		}

		next, result, err := sFnValidateConnectivityProxyCRD(context.Background(), &m)

		require.Nil(t, next)
		require.NotNil(t, result)
		require.Equal(t, time.Minute, result.RequeueAfter)
		require.Nil(t, err)
		requireContainsCondition(t, m.State.RegistryProxy.Status,
			v1alpha1.ConditionConfigured,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonConnectivityProxyCrdUnknownn,
			"Connectivity Proxy not installed. This module is required. ")
	})

	t.Run("when Connectivity Proxy CRD is installed should update condition and proceed to next state", func(t *testing.T) {
		m := fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: v1alpha1.RegistryProxy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "rp",
						Namespace: "maslo",
					},
				},
			},
			Cache: cache.NewInMemoryBoolCache(),
		}
		m.Cache.Set(true)

		next, result, err := sFnValidateConnectivityProxyCRD(context.Background(), &m)

		require.NotNil(t, next)
		requireEqualFunc(t, sFnValidateReverseProxyURL, next)
		require.Nil(t, result)
		require.Nil(t, err)
		requireContainsCondition(t, m.State.RegistryProxy.Status,
			v1alpha1.ConditionConfigured,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonConnectivityProxyCrdFound,
			"Connectivity Proxy installed.")
	})
}
