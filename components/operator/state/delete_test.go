package state

import (
	"context"
	"testing"

	"github.tools.sap/kyma/registry-proxy/components/operator/fsm"

	"github.com/stretchr/testify/require"
	"github.tools.sap/kyma/registry-proxy/components/operator/api/v1alpha1"
	"github.tools.sap/kyma/registry-proxy/components/operator/chart"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	testDeletingOperator = func() v1alpha1.RegistryProxy {
		rp := v1alpha1.RegistryProxy{
			// TypeMeta and ResourceVersion are necessary to make the fake k8s client happy
			TypeMeta: metav1.TypeMeta{
				Kind:       "RegistryProxy",
				APIVersion: v1alpha1.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:            registryProxyName,
				Namespace:       registryProxyNamespace,
				ResourceVersion: "2",
			},
			Status: v1alpha1.RegistryProxyStatus{
				State: v1alpha1.StateDeleting,
				Conditions: []metav1.Condition{
					{
						Type:   string(v1alpha1.ConditionTypeDeleted),
						Reason: string(v1alpha1.ConditionReasonDeletion),
						Status: metav1.ConditionUnknown,
					},
				},
			},
		}
		return rp
	}()
)

func Test_sFnDeleteResources(t *testing.T) {
	ns := corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-namespace"}}

	t.Run("update condition", func(t *testing.T) {
		m := &fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: v1alpha1.RegistryProxy{},
				ChartConfig: &chart.Config{
					Cache: fixManifestCache("\t"),
					CacheKey: types.NamespacedName{
						Name:      registryProxyName,
						Namespace: registryProxyNamespace,
					},
				},
			},
		}

		next, result, err := sFnDeleteResources(context.Background(), m)
		require.Nil(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnSafeDeletionState, next)

		status := m.State.RegistryProxy.Status
		require.Equal(t, v1alpha1.StateDeleting, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeDeleted,
			metav1.ConditionUnknown,
			v1alpha1.ConditionReasonDeletion,
			"Uninstalling",
		)
	})

	t.Run("safe deletion error while checking orphan resources", func(t *testing.T) {
		m := &fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: *testDeletingOperator.DeepCopy(),
				ChartConfig: &chart.Config{
					Cache: fixManifestCache("\t"),
					CacheKey: types.NamespacedName{
						Name:      registryProxyName,
						Namespace: registryProxyNamespace,
					},
				},
			},
		}
		next, result, _ := sFnSafeDeletionState(context.Background(), m)
		require.Nil(t, result)
		requireEqualFunc(t, nil, next)

		status := m.State.RegistryProxy.Status
		require.Equal(t, v1alpha1.StateWarning, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeDeleted,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonDeletionErr,
			"could not parse chart manifest: yaml: found character that cannot start any token",
		)
	})

	t.Run("safe deletion", func(t *testing.T) {
		m := &fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: *testDeletingOperator.DeepCopy(),
				ChartConfig: &chart.Config{
					Cache: fixEmptyManifestCache(),
					CacheKey: types.NamespacedName{
						Name:      registryProxyName,
						Namespace: registryProxyNamespace,
					},
					Cluster: chart.Cluster{
						Client: fake.NewClientBuilder().
							WithScheme(scheme.Scheme).
							WithObjects(&ns).
							Build(),
					},
				},
			}}
		next, result, err := sFnSafeDeletionState(context.Background(), m)
		require.Nil(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnRemoveFinalizer, next)

		status := m.State.RegistryProxy.Status
		require.Equal(t, v1alpha1.StateDeleting, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeDeleted,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonDeleted,
			"Registry Proxy module deleted",
		)
	})
}

func fixManifestCache(manifest string) chart.ManifestCache {
	cache := chart.NewInMemoryManifestCache()
	_ = cache.Set(context.Background(), types.NamespacedName{
		Name:      registryProxyName,
		Namespace: registryProxyNamespace,
	}, chart.RegistryProxySpecManifest{Manifest: manifest, CustomFlags: map[string]interface{}{
		"global": map[string]interface{}{
			"commonLabels": map[string]interface{}{
				"app.kubernetes.io/managed-by": "registry-proxy-operator",
			},
		},
		"controllerManager": map[string]interface{}{
			"container": map[string]interface{}{
				"env": map[string]interface{}{
					"ISTIO_INSTALLED": "\"false\"",
				},
			},
		},
	}})

	return cache
}

func fixEmptyManifestCache() chart.ManifestCache {
	return fixManifestCache("---")
}
