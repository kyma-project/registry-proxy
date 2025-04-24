package state

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	v1alpha2 "github.tools.sap/kyma/registry-proxy/components/registry-proxy/api/v1alpha1"
	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/fsm"
	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/resources"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func Test_sFnHandleService(t *testing.T) {
	t.Run("when service does not exist on kubernetes should create service and apply it", func(t *testing.T) {
		someService := corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "serverless",
				Namespace: "wherever",
			},
		}
		scheme := minimalScheme(t)
		updateWasCalled := false
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&someService).WithInterceptorFuncs(interceptor.Funcs{
			Update: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
				updateWasCalled = true
				return nil
			},
		}).Build()

		m := fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: v1alpha2.RegistryProxy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "rp",
						Namespace: "maslo",
					},
					Spec: v1alpha2.RegistryProxySpec{
						ProxyURL:   "http://test-proxy-url",
						TargetHost: "dummy",
					},
				},
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}
		next, result, err := sFnHandleService(context.Background(), &m)
		require.Nil(t, err)
		require.NotNil(t, result)
		require.Equal(t, ctrl.Result{RequeueAfter: time.Minute}, *result)
		require.Nil(t, next)
		require.False(t, updateWasCalled)
	})

	t.Run("when cannot get service from kubernetes should stop processing", func(t *testing.T) {
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
				RegistryProxy: v1alpha2.RegistryProxy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "rp",
						Namespace: "maslo",
					},
					Spec: v1alpha2.RegistryProxySpec{
						ProxyURL:   "http://test-proxy-url",
						TargetHost: "dummy",
					},
				},
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}
		next, result, err := sFnHandleService(context.Background(), &m)
		require.NotNil(t, err)
		require.ErrorContains(t, err, "typical error message")
		require.Nil(t, result)
		require.Nil(t, next)
		require.False(t, createOrUpdateWasCalled)
	})

	t.Run("when service does not exist on kubernetes and create fails should stop processing", func(t *testing.T) {
		scheme := minimalScheme(t)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{
			Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
				return errors.New("funny error message")
			},
		}).Build()

		m := fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: v1alpha2.RegistryProxy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "rp",
						Namespace: "maslo",
					},
					Spec: v1alpha2.RegistryProxySpec{
						ProxyURL:   "http://test-proxy-url",
						TargetHost: "dummy",
					},
				},
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}
		next, result, err := sFnHandleService(context.Background(), &m)
		require.NotNil(t, err)
		require.ErrorContains(t, err, "funny error message")
		require.Nil(t, result)
		require.Nil(t, next)
	})

	t.Run("when deployment exists on kubernetes, no changes in Service needed, and NodePort is empty, requeue", func(t *testing.T) {
		rp := v1alpha2.RegistryProxy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "rp",
				Namespace: "maslo",
			},
			Spec: v1alpha2.RegistryProxySpec{
				ProxyURL:   "http://test-proxy-url",
				TargetHost: "dummy",
			},
		}
		service := resources.NewService(&rp)
		scheme := minimalScheme(t)
		createOrUpdateWasCalled := false
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(service).WithInterceptorFuncs(interceptor.Funcs{
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
				RegistryProxy: rp,
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}
		next, result, err := sFnHandleService(context.Background(), &m)
		require.Nil(t, err)
		require.NotNil(t, result)
		require.Equal(t, ctrl.Result{RequeueAfter: time.Minute}, *result)
		require.Nil(t, next)
		require.False(t, createOrUpdateWasCalled)
		require.Empty(t, m.State.RegistryProxy.Status.Conditions)
		require.NotNil(t, m.State.Service)
	})
	t.Run("when deployment exists on kubernetes, no changes in Service needed, and NodePort is ready, update RP status", func(t *testing.T) {
		rp := v1alpha2.RegistryProxy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "rp",
				Namespace: "maslo",
			},
			Spec: v1alpha2.RegistryProxySpec{
				ProxyURL:   "http://test-proxy-url",
				TargetHost: "dummy",
			},
		}
		service := resources.NewService(&rp)
		service.Spec.Ports[0].NodePort = 1234
		scheme := minimalScheme(t)
		createOrUpdateWasCalled := false
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(service).WithInterceptorFuncs(interceptor.Funcs{
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
				RegistryProxy: rp,
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}
		next, result, err := sFnHandleService(context.Background(), &m)
		require.Nil(t, err)
		require.Nil(t, result)
		require.NotNil(t, next)
		requireEqualFunc(t, sFnHandleStatus, next)
		require.False(t, createOrUpdateWasCalled)
		require.Empty(t, m.State.RegistryProxy.Status.Conditions)
		require.NotNil(t, m.State.Service)
		require.Equal(t, int32(1234), m.State.NodePort)
	})
	t.Run("when service exists on kubernetes and we need changes should update it requeue", func(t *testing.T) {
		rp := v1alpha2.RegistryProxy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "rp",
				Namespace: "maslo",
			},
			Spec: v1alpha2.RegistryProxySpec{
				ProxyURL:   "http://test-proxy-url",
				TargetHost: "dummy",
			},
		}
		service := resources.NewService(&rp)
		service.Spec.Type = corev1.ServiceTypeClusterIP
		scheme := minimalScheme(t)
		createWasCalled := false
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(service).WithInterceptorFuncs(interceptor.Funcs{
			Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
				createWasCalled = true
				return nil
			},
		}).Build()

		m := fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: rp,
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}
		next, result, err := sFnHandleService(context.Background(), &m)
		require.Nil(t, err)
		require.NotNil(t, result)
		require.Equal(t, ctrl.Result{RequeueAfter: time.Minute}, *result)
		require.Nil(t, next)
		require.False(t, createWasCalled)
		updatedService := &corev1.Service{}
		getErr := fakeClient.Get(context.Background(), client.ObjectKey{
			Name:      "rp",
			Namespace: "maslo",
		}, updatedService)
		require.NoError(t, getErr)
		require.Equal(t, updatedService.Spec.Type, corev1.ServiceTypeNodePort)
	})

	t.Run("when deployment exists on kubernetes and update fails should stop processing", func(t *testing.T) {
		rp := v1alpha2.RegistryProxy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "rp",
				Namespace: "maslo",
			},
			Spec: v1alpha2.RegistryProxySpec{
				ProxyURL:   "http://test-proxy-url",
				TargetHost: "dummy",
			},
		}
		service := resources.NewService(&rp)
		service.Spec.Type = corev1.ServiceTypeClusterIP
		scheme := minimalScheme(t)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(service).WithInterceptorFuncs(interceptor.Funcs{
			Update: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
				return errors.New("sad error message")
			},
		}).Build()

		m := fsm.StateMachine{
			State: fsm.SystemState{
				RegistryProxy: rp,
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}
		next, result, err := sFnHandleService(context.Background(), &m)
		require.NotNil(t, err)
		require.ErrorContains(t, err, "sad error message")
		require.Nil(t, result)
		require.Nil(t, next)
	})
}
