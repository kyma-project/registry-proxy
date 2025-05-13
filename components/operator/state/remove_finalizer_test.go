package state

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.tools.sap/kyma/registry-proxy/components/operator/api/v1alpha1"
	"github.tools.sap/kyma/registry-proxy/components/operator/fsm"
	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_sFnRemoveFinalizer(t *testing.T) {
	t.Run("remove finalizer", func(t *testing.T) {
		rp := v1alpha1.RegistryProxy{
			// TypeMeta and ResourceVersion are necessary to make the fake k8s client happy
			TypeMeta: v1.TypeMeta{
				Kind:       "RegistryProxy",
				APIVersion: v1alpha1.GroupVersion.String(),
			},
			ObjectMeta: v1.ObjectMeta{
				Name:            registryProxyName,
				Namespace:       registryProxyNamespace,
				ResourceVersion: "2",
				Finalizers: []string{
					v1alpha1.Finalizer,
				},
			},
			Status: v1alpha1.RegistryProxyStatus{},
		}
		rpUnstructuredObject, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&rp)
		rpUnstructured := unstructured.Unstructured{Object: rpUnstructuredObject}
		require.NoError(t, err)

		scheme := minimalScheme(t)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&rpUnstructured).Build()
		m := fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: rp,
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}

		// remove finalizer
		next, result, err := sFnRemoveFinalizer(context.Background(), &m)
		require.Nil(t, err)
		require.Nil(t, result)
		require.Nil(t, next)
	})

	t.Run("requeue when is no finalizer", func(t *testing.T) {
		rp := v1alpha1.RegistryProxy{
			// TypeMeta and ResourceVersion are necessary to make the fake k8s client happy
			TypeMeta: v1.TypeMeta{
				Kind:       "RegistryProxy",
				APIVersion: v1alpha1.GroupVersion.String(),
			},
			ObjectMeta: v1.ObjectMeta{
				Name:            registryProxyName,
				Namespace:       registryProxyNamespace,
				ResourceVersion: "2",
			},
			Status: v1alpha1.RegistryProxyStatus{},
		}
		rpUnstructuredObject, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&rp)
		rpUnstructured := unstructured.Unstructured{Object: rpUnstructuredObject}
		require.NoError(t, err)

		scheme := minimalScheme(t)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&rpUnstructured).Build()
		m := fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: rp,
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}
		// remove finalizer
		next, result, err := sFnRemoveFinalizer(context.Background(), &m)
		require.Nil(t, err)
		require.Equal(t, &ctrl.Result{Requeue: true}, result)
		require.Nil(t, next)
	})
}
