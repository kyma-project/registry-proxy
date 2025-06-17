package state

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.tools.sap/kyma/registry-proxy/components/operator/api/v1alpha1"
	"github.tools.sap/kyma/registry-proxy/components/operator/fsm"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_sFnServedFilter(t *testing.T) {
	t.Run("skip processing when served is false", func(t *testing.T) {
		m := fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: v1alpha1.RegistryProxy{
					Status: v1alpha1.RegistryProxyStatus{
						Served: v1alpha1.ServedFalse,
					},
				},
			},
		}

		nextFn, result, err := sFnServedFilter(context.TODO(), &m)
		require.Nil(t, err)
		require.Nil(t, result)
		require.Nil(t, nextFn)
	})

	t.Run("do next step when served is true", func(t *testing.T) {
		m := fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: v1alpha1.RegistryProxy{
					Status: v1alpha1.RegistryProxyStatus{
						Served: v1alpha1.ServedTrue,
					},
				},
			},
		}

		nextFn, result, err := sFnServedFilter(context.TODO(), &m)
		require.Nil(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnValidateConnectivityProxyCRD, nextFn)
	})

	t.Run("set served value from nil to true when there is no served registry proxy on cluster", func(t *testing.T) {

		scheme := minimalScheme(t)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		m := fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: v1alpha1.RegistryProxy{
					Status: v1alpha1.RegistryProxyStatus{},
				},
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}

		next, result, err := sFnServedFilter(context.TODO(), &m)
		require.Nil(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnValidateConnectivityProxyCRD, next)
		require.Equal(t, v1alpha1.ServedTrue, m.State.RegistryProxy.Status.Served)
	})

	t.Run("set served value from nil to false and set condition to error when there is at lease one served registry proxy on cluster", func(t *testing.T) {

		existingRP := minimalRegistryProxy()

		scheme := minimalScheme(t)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(existingRP).Build()
		m := fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: v1alpha1.RegistryProxy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "new-registry-proxy",
						Namespace: "kyma-system",
					},
					Status: v1alpha1.RegistryProxyStatus{},
				},
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}

		next, result, err := sFnServedFilter(context.TODO(), &m)

		expectedErrorMessage := "Only one instance of RegistryProxy is allowed (current served instance: kyma-system/existing-registry-proxy). This RegistryProxy CR is redundant. Remove it to fix the problem."
		require.EqualError(t, err, expectedErrorMessage)
		require.Nil(t, result)
		require.Nil(t, next)
		require.Equal(t, v1alpha1.ServedFalse, m.State.RegistryProxy.Status.Served)

		require.Equal(t, v1alpha1.StateWarning, m.State.RegistryProxy.Status.State)
		requireContainsCondition(t, m.State.RegistryProxy.Status,
			v1alpha1.ConditionTypeConfigured,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonRegistryProxyDuplicated,
			expectedErrorMessage,
		)
	})
}

// TODO: export to common test file?
func minimalRegistryProxy() *unstructured.Unstructured {
	registryProxy := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.kyma-project.io/v1alpha1",
			"kind":       "RegistryProxy",
			"metadata": map[string]interface{}{
				"name":      "existing-registry-proxy",
				"namespace": "kyma-system",
			},
			"spec": map[string]interface{}{},
			"status": map[string]interface{}{
				"served": "True",
			},
		},
	}
	return registryProxy
}
