package state

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/api/v1alpha1"
	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/fsm"
	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/resources"
	"go.uber.org/zap"
	apisecurityv1 "istio.io/api/security/v1"
	securityclientv1 "istio.io/client-go/pkg/apis/security/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func Test_sFnHandlePeerAuthentication(t *testing.T) {
	os.Setenv("ISTIO_INSTALLED", "true")
	t.Run("when PeerAuthentication does not exist on kubernetes should create PeerAuthentication and apply it", func(t *testing.T) {
		somePeerAuthentication := securityclientv1.PeerAuthentication{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "serversome-peer",
				Namespace: "wherever",
			},
		}
		scheme := minimalScheme(t)
		updateWasCalled := false

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&somePeerAuthentication).WithInterceptorFuncs(interceptor.Funcs{
			Update: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
				updateWasCalled = true
				return nil
			},
		}).Build()

		m := fsm.StateMachine{
			State: fsm.SystemState{
				Connection: v1alpha1.Connection{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "connection",
						Namespace: "maslo",
					},
					Spec: v1alpha1.ConnectionSpec{
						Proxy: v1alpha1.ConnectionSpecProxy{
							URL: "http://test-proxy-url",
						},
						Target: v1alpha1.ConnectionSpecTarget{
							Host: "dummy",
						},
					},
				},
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}
		next, result, err := sFnHandlePeerAuthentication(context.Background(), &m)
		require.Nil(t, err)
		require.NotNil(t, result)
		require.Equal(t, ctrl.Result{RequeueAfter: time.Minute}, *result)
		require.Nil(t, next)
		require.False(t, updateWasCalled)
	})

	t.Run("when cannot get PeerAuthentication from kubernetes should stop processing", func(t *testing.T) {
		scheme := minimalScheme(t)
		createOrUpdateWasCalled := false
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{
			Get: func(ctx context.Context, client client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
				return errors.New("typical error message")
			},
			Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
				createOrUpdateWasCalled = true
				return nil
			},
			Update: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
				createOrUpdateWasCalled = true
				return nil
			},
		}).Build()

		m := fsm.StateMachine{
			State: fsm.SystemState{
				Connection: v1alpha1.Connection{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "connection",
						Namespace: "maslo",
					},
					Spec: v1alpha1.ConnectionSpec{
						Proxy: v1alpha1.ConnectionSpecProxy{
							URL: "http://test-proxy-url",
						},
						Target: v1alpha1.ConnectionSpecTarget{
							Host: "dummy",
						},
					},
				},
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}
		next, result, err := sFnHandlePeerAuthentication(context.Background(), &m)
		require.NotNil(t, err)
		require.ErrorContains(t, err, "typical error message")
		require.Nil(t, result)
		require.Nil(t, next)
		require.False(t, createOrUpdateWasCalled)
	})
	t.Run("when PeerAuthentication does not exist on kubernetes and create fails should stop processing", func(t *testing.T) {
		scheme := minimalScheme(t)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{
			Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
				return errors.New("funny error message")
			},
		}).Build()

		m := fsm.StateMachine{
			State: fsm.SystemState{
				Connection: v1alpha1.Connection{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "connection",
						Namespace: "maslo",
					},
					Spec: v1alpha1.ConnectionSpec{
						Proxy: v1alpha1.ConnectionSpecProxy{
							URL: "http://test-proxy-url",
						},
						Target: v1alpha1.ConnectionSpecTarget{
							Host: "dummy",
						},
					},
				},
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}
		next, result, err := sFnHandlePeerAuthentication(context.Background(), &m)
		require.NotNil(t, err)
		require.ErrorContains(t, err, "funny error message")
		require.Nil(t, result)
		require.Nil(t, next)
	})

	t.Run("when PeerAuthentication exists on kubernetes and we need changes should update it requeue", func(t *testing.T) {
		connection := v1alpha1.Connection{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "connection",
				Namespace: "maslo",
			},
			Spec: v1alpha1.ConnectionSpec{
				Proxy: v1alpha1.ConnectionSpecProxy{
					URL: "http://test-proxy-url",
				},
				Target: v1alpha1.ConnectionSpecTarget{
					Host: "dummy",
				},
			},
		}
		pa := resources.NewPeerAuthentication(&connection)
		pa.Spec.Mtls.Mode = apisecurityv1.PeerAuthentication_MutualTLS_STRICT
		scheme := minimalScheme(t)
		createWasCalled := false
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pa).WithInterceptorFuncs(interceptor.Funcs{
			Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
				createWasCalled = true
				return nil
			},
		}).Build()

		m := fsm.StateMachine{
			State: fsm.SystemState{
				Connection: connection,
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}
		next, result, err := sFnHandlePeerAuthentication(context.Background(), &m)
		require.Nil(t, err)
		require.NotNil(t, result)
		require.Equal(t, ctrl.Result{RequeueAfter: time.Minute}, *result)
		require.Nil(t, next)
		require.False(t, createWasCalled)
		updatedPA := &securityclientv1.PeerAuthentication{}
		getErr := fakeClient.Get(context.Background(), client.ObjectKey{
			Name:      "connection",
			Namespace: "maslo",
		}, updatedPA)
		require.NoError(t, getErr)
		require.Equal(t, updatedPA.Spec.Mtls.Mode, apisecurityv1.PeerAuthentication_MutualTLS_PERMISSIVE)
	})
	t.Run("when PeerAuthentication exists on kubernetes and update fails should stop processing", func(t *testing.T) {
		connection := v1alpha1.Connection{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "connection",
				Namespace: "maslo",
			},
			Spec: v1alpha1.ConnectionSpec{
				Proxy: v1alpha1.ConnectionSpecProxy{
					URL: "http://test-proxy-url",
				},
				Target: v1alpha1.ConnectionSpecTarget{
					Host: "dummy",
				},
			},
		}
		pa := resources.NewPeerAuthentication(&connection)
		pa.Spec.Mtls.Mode = apisecurityv1.PeerAuthentication_MutualTLS_STRICT
		scheme := minimalScheme(t)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pa).WithInterceptorFuncs(interceptor.Funcs{
			Update: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
				return errors.New("sad error message")
			},
		}).Build()

		m := fsm.StateMachine{
			State: fsm.SystemState{
				Connection: connection,
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}
		next, result, err := sFnHandlePeerAuthentication(context.Background(), &m)
		require.NotNil(t, err)
		require.ErrorContains(t, err, "sad error message")
		require.Nil(t, result)
		require.Nil(t, next)
	})
}
