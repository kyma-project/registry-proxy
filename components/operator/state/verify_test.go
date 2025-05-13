package state

import (
	"context"
	"github.tools.sap/kyma/registry-proxy/components/operator/fsm"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.tools.sap/kyma/registry-proxy/components/operator/api/v1alpha1"
	"github.tools.sap/kyma/registry-proxy/components/operator/chart"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	testInstalledRegistryProxy = func() v1alpha1.RegistryProxy {
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
				Finalizers: []string{
					v1alpha1.Finalizer,
				},
			},
			Status: v1alpha1.RegistryProxyStatus{
				State: v1alpha1.StateReady,
				Conditions: []metav1.Condition{
					{
						Type:   string(v1alpha1.ConditionTypeConfigured),
						Status: metav1.ConditionTrue,
						Reason: string(v1alpha1.ConditionReasonConfiguration),
					},
					{
						Type:   string(v1alpha1.ConditionTypeInstalled),
						Status: metav1.ConditionTrue,
						Reason: string(v1alpha1.ConditionReasonInstallation),
					},
				},
			},
		}
		return rp
	}()
	testDeployCR = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deploy",
			Namespace: "default",
		},
		Status: appsv1.DeploymentStatus{
			Conditions: []appsv1.DeploymentCondition{
				{
					Type:   appsv1.DeploymentAvailable,
					Status: corev1.ConditionUnknown,
				},
			},
		},
	}
)

const (
	testDeployManifest = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deploy
  namespace: default
`
)

func Test_sFnVerifyResources(t *testing.T) {
	t.Run("ready", func(t *testing.T) {
		rpUnstructuredObject, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&testInstalledRegistryProxy)
		rpUnstructured := unstructured.Unstructured{Object: rpUnstructuredObject}
		require.NoError(t, err)
		scheme := minimalScheme(t)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&rpUnstructured).Build()

		m := &fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: *testInstalledRegistryProxy.DeepCopy(),
				ChartConfig: &chart.Config{
					Cache: fixEmptyManifestCache(),
					CacheKey: types.NamespacedName{
						Name:      testInstalledRegistryProxy.GetName(),
						Namespace: testInstalledRegistryProxy.GetNamespace(),
					},
				},
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}

		// verify and return update condition state
		next, result, err := sFnVerifyResources(context.Background(), m)
		require.Nil(t, err)
		require.Nil(t, result)
		require.Nil(t, next)

		status := m.State.RegistryProxy.Status
		require.Equal(t, v1alpha1.StateReady, status.State)
		require.Len(t, status.Conditions, 2)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeInstalled,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonInstalled,
			"Registry Proxy installed",
		)
	})

	t.Run("verify error", func(t *testing.T) {
		rpUnstructuredObject, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&testInstalledRegistryProxy)
		rpUnstructured := unstructured.Unstructured{Object: rpUnstructuredObject}
		require.NoError(t, err)
		scheme := minimalScheme(t)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&rpUnstructured).Build()

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
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}

		// handle verify err and update condition with err
		next, result, err := sFnVerifyResources(context.Background(), m)
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

	t.Run("requeue when resources are not ready", func(t *testing.T) {
		scheme := minimalScheme(t)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(testDeployCR).Build()

		m := &fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: *testInstalledRegistryProxy.DeepCopy(),
				ChartConfig: &chart.Config{
					Cache: func() chart.ManifestCache {
						cache := chart.NewInMemoryManifestCache()
						_ = cache.Set(context.Background(), types.NamespacedName{
							Name:      testInstalledRegistryProxy.GetName(),
							Namespace: testInstalledRegistryProxy.GetNamespace(),
						}, chart.RegistryProxySpecManifest{Manifest: testDeployManifest})
						return cache
					}(), CacheKey: types.NamespacedName{
						Name:      testInstalledRegistryProxy.GetName(),
						Namespace: testInstalledRegistryProxy.GetNamespace(),
					},
					Cluster: chart.Cluster{
						Client: fakeClient,
					},
				},
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}

		// return requeue on verification failed
		next, result, err := sFnVerifyResources(context.Background(), m)

		_, expectedResult, _ := requeueAfter(time.Second * 3)
		require.NoError(t, err)
		require.Equal(t, expectedResult, result)
		require.Nil(t, next)
	})
}
