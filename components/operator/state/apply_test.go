package state

import (
	"context"
	"testing"

	"github.tools.sap/kyma/registry-proxy/components/operator/fsm"
	"go.uber.org/zap"

	"github.com/stretchr/testify/require"
	"github.tools.sap/kyma/registry-proxy/components/operator/api/v1alpha1"
	"github.tools.sap/kyma/registry-proxy/components/operator/chart"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func Test_buildSFnApplyResources(t *testing.T) {
	t.Run("switch state and add condition when condition is missing", func(t *testing.T) {
		m := &fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: v1alpha1.RegistryProxy{},
				ChartConfig: &chart.Config{
					Cache: fixEmptyManifestCache(),
					CacheKey: types.NamespacedName{
						Name:      testInstalledRegistryProxy.GetName(),
						Namespace: testInstalledRegistryProxy.GetNamespace(),
					},
					Release: chart.Release{
						Name:      testInstalledRegistryProxy.GetName(),
						Namespace: testInstalledRegistryProxy.GetNamespace(),
					},
				},
			},
			Log: zap.NewNop().Sugar(),
		}

		next, result, err := sFnApplyResources(context.Background(), m)
		require.Nil(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnVerifyResources, next)

		status := m.State.RegistryProxy.Status
		require.Equal(t, v1alpha1.StateProcessing, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeInstalled,
			metav1.ConditionUnknown,
			v1alpha1.ConditionReasonInstallation,
			"Installing for configuration",
		)
	})

	t.Run("apply resources", func(t *testing.T) {
		m := &fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: *testInstalledRegistryProxy.DeepCopy(),
				ChartConfig: &chart.Config{
					Cache: fixEmptyManifestCache(),
					CacheKey: types.NamespacedName{
						Name:      testInstalledRegistryProxy.GetName(),
						Namespace: testInstalledRegistryProxy.GetNamespace(),
					},
					Release: chart.Release{
						Name:      testInstalledRegistryProxy.GetName(),
						Namespace: testInstalledRegistryProxy.GetNamespace(),
					},
				},
			},
			Log: zap.NewNop().Sugar(),
		}

		// run installation process and return verificating state
		next, result, err := sFnApplyResources(context.Background(), m)
		require.Nil(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnVerifyResources, next)
	})

	t.Run("install chart error", func(t *testing.T) {
		m := &fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: *testInstalledRegistryProxy.DeepCopy(),
				ChartConfig: &chart.Config{
					Cache: fixManifestCache("\t"),
					CacheKey: types.NamespacedName{
						Name:      testInstalledRegistryProxy.GetName(),
						Namespace: testInstalledRegistryProxy.GetNamespace(),
					},
				},
			},
			Log: zap.NewNop().Sugar(),
		}

		// handle error and return update condition state
		next, result, err := sFnApplyResources(context.Background(), m)
		require.EqualError(t, err, "could not parse chart manifest: yaml: found character that cannot start any token")
		require.Nil(t, result)
		require.Nil(t, next)

		status := m.State.RegistryProxy.Status
		require.Equal(t, v1alpha1.StateError, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeInstalled,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonInstallationErr,
			"could not parse chart manifest: yaml: found character that cannot start any token",
		)
	})
}
