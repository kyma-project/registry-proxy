package state

import (
	"context"
	"testing"

	"github.com/kyma-project/registry-proxy/components/operator/api/v1alpha1"
	"github.com/kyma-project/registry-proxy/components/operator/fsm"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	registryProxyName      = "registry-proxy"
	registryProxyNamespace = "registry-proxy-namespace"
)

func Test_sFnAddFinalizer(t *testing.T) {
	t.Run("set finalizer", func(t *testing.T) {
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

		// set finalizer
		next, result, err := sFnAddFinalizer(context.Background(), &m)
		require.NoError(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnInitialize, next)

		// check finalizer in systemState
		require.Contains(t, m.State.RegistryProxy.GetFinalizers(), v1alpha1.Finalizer)

		// check finalizer in k8s
		obj := v1alpha1.RegistryProxy{}
		err = m.Client.Get(context.Background(),
			client.ObjectKey{
				Namespace: registryProxyNamespace,
				Name:      registryProxyName,
			},
			&obj)
		require.NoError(t, err)
		require.Contains(t, obj.GetFinalizers(), v1alpha1.Finalizer)
	})

	t.Run("stop when no finalizer and instance is being deleted", func(t *testing.T) {
		metaTimeNow := v1.Now()

		scheme := minimalScheme(t)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		m := fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: v1alpha1.RegistryProxy{
					ObjectMeta: v1.ObjectMeta{
						Name:              registryProxyName,
						Namespace:         registryProxyNamespace,
						DeletionTimestamp: &metaTimeNow,
					},
					Status: v1alpha1.RegistryProxyStatus{},
				},
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}

		// stop
		next, result, err := sFnAddFinalizer(context.Background(), &m)
		require.Nil(t, err)
		require.Nil(t, result)
		require.Nil(t, next)
	})
}
