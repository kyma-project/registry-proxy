package state

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/api/v1alpha1"
	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/fsm"
	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/resources"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func Test_sFnHandleDeployment(t *testing.T) {
	t.Run("when deployment does not exist on kubernetes should create deployment and apply it", func(t *testing.T) {
		someDeployment := appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "some-deployment",
				Namespace: "wherever",
			},
		}
		scheme := minimalScheme(t)
		updateWasCalled := false
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&someDeployment).WithInterceptorFuncs(interceptor.Funcs{
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
						ProxyURL:   "http://test-proxy-url",
						TargetHost: "dummy",
					},
				},
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}
		next, result, err := sFnHandleDeployment(context.Background(), &m)

		require.Nil(t, err)
		require.NotNil(t, result)
		require.Equal(t, ctrl.Result{RequeueAfter: time.Minute}, *result)
		require.Nil(t, next)
		require.False(t, updateWasCalled)

		requireContainsCondition(t, m.State.Connection.Status,
			v1alpha1.ConditionConnectionDeployed,
			metav1.ConditionUnknown,
			v1alpha1.ConditionReasonDeploymentCreated,
			"Deployment connection created")

		appliedDeployment := &appsv1.Deployment{}
		getErr := fakeClient.Get(context.Background(), client.ObjectKey{
			Name:      "connection",
			Namespace: "maslo",
		}, appliedDeployment)
		require.NoError(t, getErr)

		require.NotEmpty(t, appliedDeployment.OwnerReferences)
		require.Equal(t, "Connection", appliedDeployment.OwnerReferences[0].Kind)
		require.Equal(t, "connection", appliedDeployment.OwnerReferences[0].Name)
	})
	t.Run("when cannot get deployment from kubernetes should stop processing", func(t *testing.T) {
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
						ProxyURL:   "http://test-proxy-url",
						TargetHost: "dummy",
					},
				},
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}
		next, result, err := sFnHandleDeployment(context.Background(), &m)

		require.NotNil(t, err)
		require.ErrorContains(t, err, "typical error message")
		require.Nil(t, result)
		require.Nil(t, next)
		require.False(t, createOrUpdateWasCalled)

	})
	t.Run("when deployment does not exist on kubernetes and create fails should stop processing", func(t *testing.T) {
		scheme := minimalScheme(t)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects().WithInterceptorFuncs(interceptor.Funcs{
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
						ProxyURL:   "http://test-proxy-url",
						TargetHost: "dummy",
					},
				},
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}
		next, result, err := sFnHandleDeployment(context.Background(), &m)

		require.NotNil(t, err)
		require.ErrorContains(t, err, "funny error message")
		require.Nil(t, result)
		require.Nil(t, next)
		requireContainsCondition(t, m.State.Connection.Status,
			v1alpha1.ConditionConnectionDeployed,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonDeploymentFailed,
			"Deployment connection create failed: funny error message")
	})
	t.Run("when deployment exists on kubernetes but we do not need changes should keep it without changes and go to the next state", func(t *testing.T) {
		connection := v1alpha1.Connection{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "connection",
				Namespace: "maslo",
			},
			Spec: v1alpha1.ConnectionSpec{
				ProxyURL:   "http://test-proxy-url",
				TargetHost: "dummy",
			},
		}
		deployment := resources.NewDeployment(&connection, connection.Spec.ProxyURL)
		scheme := minimalScheme(t)
		createOrUpdateWasCalled := false
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(deployment).WithInterceptorFuncs(interceptor.Funcs{
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
				Connection: connection,
				ProxyURL:   connection.Spec.ProxyURL,
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}

		next, result, err := sFnHandleDeployment(context.Background(), &m)

		require.Nil(t, err)
		require.Nil(t, result)
		require.NotNil(t, next)
		requireEqualFunc(t, sFnHandlePodStatus, next)
		require.False(t, createOrUpdateWasCalled)
		require.Empty(t, m.State.Connection.Status.Conditions)
		require.NotNil(t, m.State.Deployment)
	})
	t.Run("when deployment exists on kubernetes and we need changes should update it and go to the next state", func(t *testing.T) {
		connection := v1alpha1.Connection{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "connection",
				Namespace: "maslo",
			},
			Spec: v1alpha1.ConnectionSpec{
				ProxyURL:   "http://test-proxy-url",
				TargetHost: "dummy",
			},
		}
		deployment := resources.NewDeployment(&connection, connection.Spec.ProxyURL)
		scheme := minimalScheme(t)
		createWasCalled := false
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(deployment).WithInterceptorFuncs(interceptor.Funcs{
			Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
				createWasCalled = true
				return nil
			},
		}).Build()

		connection.Spec.TargetHost = "fresh"

		m := fsm.StateMachine{
			State: fsm.SystemState{
				Connection: connection,
				ProxyURL:   connection.Spec.ProxyURL,
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}

		next, result, err := sFnHandleDeployment(context.Background(), &m)

		require.Nil(t, err)
		require.NotNil(t, result)
		require.Equal(t, ctrl.Result{RequeueAfter: time.Minute}, *result)
		require.Nil(t, next)
		requireContainsCondition(t, m.State.Connection.Status,
			v1alpha1.ConditionConnectionDeployed,
			metav1.ConditionUnknown,
			v1alpha1.ConditionReasonDeploymentUpdated,
			"Deployment connection updated")
		require.False(t, createWasCalled)
		updatedDeployment := &appsv1.Deployment{}
		getErr := fakeClient.Get(context.Background(), client.ObjectKey{
			Name:      "connection",
			Namespace: "maslo",
		}, updatedDeployment)
		require.NoError(t, getErr)
		// deployment should have updated some specific fields
		require.Contains(t, updatedDeployment.Spec.Template.Spec.Containers[0].Env,
			corev1.EnvVar{Name: "TARGET_HOST", Value: "fresh"})
	})
	t.Run("when deployment exists on kubernetes and update fails should stop processing", func(t *testing.T) {
		connection := v1alpha1.Connection{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "connection",
				Namespace: "maslo",
			},
			Spec: v1alpha1.ConnectionSpec{
				ProxyURL:   "http://test-proxy-url",
				TargetHost: "dummy",
			},
		}
		deployment := resources.NewDeployment(&connection, connection.Spec.ProxyURL)
		scheme := minimalScheme(t)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(deployment).WithInterceptorFuncs(interceptor.Funcs{
			Update: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
				return errors.New("sad error message")
			},
		}).Build()

		connection.Spec.TargetHost = "fresh"

		m := fsm.StateMachine{
			State: fsm.SystemState{
				Connection: connection,
				ProxyURL:   connection.Spec.ProxyURL,
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}

		next, result, err := sFnHandleDeployment(context.Background(), &m)
		require.NotNil(t, err)
		require.ErrorContains(t, err, "sad error message")
		require.Nil(t, result)
		require.Nil(t, next)
		requireContainsCondition(t, m.State.Connection.Status,
			v1alpha1.ConditionConnectionDeployed,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonDeploymentFailed,
			"Deployment connection update failed: sad error message")
	})
}
