package state

import (
	"context"
	"testing"

	"github.com/kyma-project/registry-proxy/components/operator/api/v1alpha1"
	"github.com/kyma-project/registry-proxy/components/operator/fsm"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_sFnInitialize(t *testing.T) {
	t.Run("setup and return next step sFnApplyResources", func(t *testing.T) {
		scheme := minimalScheme(t)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		m := fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: v1alpha1.RegistryProxy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      registryProxyName,
						Namespace: registryProxyNamespace,
						Finalizers: []string{
							v1alpha1.Finalizer,
						},
					},
					Status: v1alpha1.RegistryProxyStatus{},
				},
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}

		next, result, err := sFnInitialize(context.Background(), &m)
		require.Nil(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnValidateConnectivityProxyCRD, next)
	})

	t.Run("setup and return next step sFnDeleteResources", func(t *testing.T) {
		scheme := minimalScheme(t)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		metaTime := metav1.Now()
		m := fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: v1alpha1.RegistryProxy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      registryProxyName,
						Namespace: registryProxyNamespace,
						Finalizers: []string{
							v1alpha1.Finalizer,
						},
						DeletionTimestamp: &metaTime,
					},
					Status: v1alpha1.RegistryProxyStatus{},
				},
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}

		next, result, err := sFnInitialize(context.Background(), &m)
		require.Nil(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnDeleteResources, next)
	})
}
