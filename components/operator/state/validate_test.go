package state

import (
	"context"

	"github.com/kyma-project/registry-proxy/components/common/cache"
	"github.com/kyma-project/registry-proxy/components/operator/api/v1alpha1"
	"github.com/kyma-project/registry-proxy/components/operator/fsm"
	"github.com/stretchr/testify/require"

	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_sFnValidateConnectivityProxyCRD(t *testing.T) {
	t.Run("when Connectivity Proxy CRD is not installed should update condition and requeue", func(t *testing.T) {
		m := fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: *testInstalledRegistryProxy.DeepCopy(),
			},
			ConnectivityProxyReadiness: cache.NewInMemoryBoolCache(),
		}

		next, result, err := sFnValidateConnectivityProxyCRD(context.Background(), &m)

		require.Nil(t, next)
		require.NotNil(t, result)
		require.Equal(t, time.Minute, result.RequeueAfter)
		require.Nil(t, err)
		requireContainsCondition(t, m.State.RegistryProxy.Status,
			v1alpha1.ConditionPrerequisitesSatisfied,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonConnectivityProxyUnavailable,
			"Connectivity Proxy is unavailable. This module is required.")
		require.Equal(t, v1alpha1.StateWarning, m.State.RegistryProxy.Status.State,
			"State should be set to Warning when prerequisites are not satisfied.")
	})

	t.Run("when Connectivity Proxy CRD is not installed but spec proxy is set should update condition and proceed to next state", func(t *testing.T) {
		m := fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: *testInstalledRegistryProxy.DeepCopy(),
			},
			ConnectivityProxyReadiness: cache.NewInMemoryBoolCache(),
		}
		m.State.RegistryProxy.Spec.Proxy.URL = "http://my-proxy.com"

		next, result, err := sFnValidateConnectivityProxyCRD(context.Background(), &m)

		require.NotNil(t, next)
		requireEqualFunc(t, sFnApplyResources, next)
		require.Nil(t, result)
		require.Nil(t, err)
		requireContainsCondition(t, m.State.RegistryProxy.Status,
			v1alpha1.ConditionPrerequisitesSatisfied,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonConnectivityProxySkipped,
			"Connectivity Proxy check skipped, .spec.proxy.url is set.")
	})

	t.Run("when Connectivity Proxy CRD is installed should update condition and proceed to next state", func(t *testing.T) {
		m := fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: *testInstalledRegistryProxy.DeepCopy(),
			},
			ConnectivityProxyReadiness: cache.NewInMemoryBoolCache(),
		}
		m.ConnectivityProxyReadiness.Set(true)

		next, result, err := sFnValidateConnectivityProxyCRD(context.Background(), &m)

		require.NotNil(t, next)
		requireEqualFunc(t, sFnApplyResources, next)
		require.Nil(t, result)
		require.Nil(t, err)
		requireContainsCondition(t, m.State.RegistryProxy.Status,
			v1alpha1.ConditionPrerequisitesSatisfied,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonConnectivityProxyAvailable,
			"Connectivity Proxy installed.")
	})
}
