package state

import (
	"context"

	"github.com/stretchr/testify/require"
	"github.tools.sap/kyma/registry-proxy/components/common/cache"
	"github.tools.sap/kyma/registry-proxy/components/operator/api/v1alpha1"
	"github.tools.sap/kyma/registry-proxy/components/operator/fsm"

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
			Cache: cache.NewInMemoryBoolCache(),
		}

		next, result, err := sFnValidateConnectivityProxyCRD(context.Background(), &m)

		require.Nil(t, next)
		require.NotNil(t, result)
		require.Equal(t, time.Minute, result.RequeueAfter)
		require.Nil(t, err)
		requireContainsCondition(t, m.State.RegistryProxy.Status,
			v1alpha1.ConditionPrerequisitesSatisfied,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonConnectivityProxyCrdUnknown,
			"Connectivity Proxy not installed. This module is required.")
		require.Equal(t, v1alpha1.StateWarning, m.State.RegistryProxy.Status.State,
			"State should be set to Warning when prerequisites are not satisfied.")
	})

	t.Run("when Connectivity Proxy CRD is installed should update condition and proceed to next state", func(t *testing.T) {
		m := fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: *testInstalledRegistryProxy.DeepCopy(),
			},
			Cache: cache.NewInMemoryBoolCache(),
		}
		m.Cache.Set(true)

		next, result, err := sFnValidateConnectivityProxyCRD(context.Background(), &m)

		require.NotNil(t, next)
		requireEqualFunc(t, sFnApplyResources, next)
		require.Nil(t, result)
		require.Nil(t, err)
		requireContainsCondition(t, m.State.RegistryProxy.Status,
			v1alpha1.ConditionPrerequisitesSatisfied,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonConnectivityProxyCrdFound,
			"Connectivity Proxy installed.")
	})
}
