package state

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-project/registry-proxy/components/registry-proxy/api/v1alpha1"
	"github.com/kyma-project/registry-proxy/components/registry-proxy/fsm"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_sFnHandlePodStatus(t *testing.T) {
	t.Run("no pod exists", func(t *testing.T) {
		somePod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "some-pod",
				Namespace: "wherever",
			},
		}
		scheme := minimalScheme(t)

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(somePod).Build()

		m := fsm.StateMachine{
			State: fsm.SystemState{
				Connection: v1alpha1.Connection{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "connection",
						Namespace: "maslo",
					},
				},
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}

		next, result, err := sFnHandlePodStatus(context.Background(), &m)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, ctrl.Result{RequeueAfter: time.Minute}, *result)
		require.Nil(t, next)
		requireContainsCondition(t, m.State.Connection.Status, v1alpha1.ConditionConnectionDeployed, metav1.ConditionFalse, v1alpha1.ConditionReasonResourcesNotReady, "no pod exists")
		requireContainsCondition(t, m.State.Connection.Status, v1alpha1.ConditionConnectionReady, metav1.ConditionFalse, v1alpha1.ConditionReasonNotEstabilished, "no pod exists")
	})

	t.Run("one pod exists", func(t *testing.T) {
		onePod := minimalPod(true)
		onePod.CreationTimestamp = metav1.NewTime(time.Now().Add(-time.Hour))
		scheme := minimalScheme(t)

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(onePod).Build()

		m := fsm.StateMachine{
			State: fsm.SystemState{
				Connection: v1alpha1.Connection{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "connection",
						Namespace: "maslo",
					},
				},
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}

		next, result, err := sFnHandlePodStatus(context.Background(), &m)
		require.NoError(t, err)
		require.Nil(t, result)
		require.NotNil(t, next)
		requireEqualFunc(t, sFnHandlePeerAuthentication, next)
	})
	t.Run("multiple pods exist", func(t *testing.T) {
		// create two pods with different creation timestamps, the first has failing probes
		firstPod := minimalPod(false)
		firstPod.CreationTimestamp = metav1.NewTime(time.Now().Add(-time.Hour))
		secondPod := minimalPod(true)
		secondPod.Name = "rp-pod2"
		secondPod.CreationTimestamp = metav1.NewTime(time.Now().Add(-time.Minute))
		scheme := minimalScheme(t)

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(firstPod, secondPod).Build()

		m := fsm.StateMachine{
			State: fsm.SystemState{
				Connection: v1alpha1.Connection{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "connection",
						Namespace: "maslo",
					},
				},
			},
			Log:    zap.NewNop().Sugar(),
			Client: fakeClient,
			Scheme: scheme,
		}

		next, result, err := sFnHandlePodStatus(context.Background(), &m)
		require.NoError(t, err)
		require.Nil(t, result)
		require.NotNil(t, next)
		requireEqualFunc(t, sFnHandlePeerAuthentication, next)
		requireContainsCondition(t, m.State.Connection.Status, v1alpha1.ConditionConnectionDeployed, metav1.ConditionTrue, v1alpha1.ConditionReasonResourcesDeployed, "Reverse-proxy ready")
	})
}

func TestHandleProbe(t *testing.T) {
	tests := []struct {
		name              string
		rp                *v1alpha1.Connection
		phase             corev1.PodPhase
		expectedCondition metav1.Condition
	}{
		{
			name:  "should return error not ready",
			rp:    &v1alpha1.Connection{},
			phase: corev1.PodPending,
			expectedCondition: metav1.Condition{
				Type:    string(v1alpha1.ConditionConnectionDeployed),
				Status:  metav1.ConditionFalse,
				Reason:  string(v1alpha1.ConditionReasonError),
				Message: "Reverse-proxy not ready: pod is in phase Pending",
			},
		},
		{
			name:  "should return success",
			rp:    &v1alpha1.Connection{},
			phase: corev1.PodRunning,
			expectedCondition: metav1.Condition{
				Type:    string(v1alpha1.ConditionConnectionDeployed),
				Status:  metav1.ConditionTrue,
				Reason:  string(v1alpha1.ConditionReasonResourcesDeployed),
				Message: "Reverse-proxy ready",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = handleLivenessStatus(tt.rp, tt.phase)
			requireContainsCondition(t, tt.rp.Status,
				v1alpha1.ConditionType(tt.expectedCondition.Type),
				tt.expectedCondition.Status,
				v1alpha1.ConditionReason(tt.expectedCondition.Reason),
				tt.expectedCondition.Message,
			)
		})
	}
}

func TestHandleReadinessStatus(t *testing.T) {
	tests := []struct {
		name              string
		rp                *v1alpha1.Connection
		conditions        []corev1.PodCondition
		expectedCondition metav1.Condition
	}{
		{
			name: "should return error on unready pod",
			rp:   &v1alpha1.Connection{},
			conditions: []corev1.PodCondition{
				{
					Type:    corev1.PodReady,
					Status:  corev1.ConditionFalse,
					Reason:  "ContainersNotReady",
					Message: "containers with unready status: [a]",
				},
			},
			expectedCondition: metav1.Condition{
				Type:    string(v1alpha1.ConditionConnectionReady),
				Status:  metav1.ConditionFalse,
				Reason:  string(v1alpha1.ConditionReasonNotEstabilished),
				Message: "Target registry not reachable: ContainersNotReady",
			},
		},
		{
			name:       "should return error on missing condition",
			rp:         &v1alpha1.Connection{},
			conditions: []corev1.PodCondition{},
			expectedCondition: metav1.Condition{
				Type:    string(v1alpha1.ConditionConnectionReady),
				Status:  metav1.ConditionFalse,
				Reason:  string(v1alpha1.ConditionReasonNotEstabilished),
				Message: "Target registry not reachable: no condition found",
			},
		},
		{
			name: "should return success",
			rp:   &v1alpha1.Connection{},
			conditions: []corev1.PodCondition{
				{
					Type:    corev1.PodReady,
					Status:  corev1.ConditionTrue,
					Reason:  "ContainersReady",
					Message: "all containers in pod are ready",
				},
			},
			expectedCondition: metav1.Condition{
				Type:    string(v1alpha1.ConditionConnectionReady),
				Status:  metav1.ConditionTrue,
				Reason:  string(v1alpha1.ConditionReasonEstabilished),
				Message: "Target registry reachable",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = handleReadinessStatus(tt.rp, tt.conditions)
			requireContainsCondition(t, tt.rp.Status,
				v1alpha1.ConditionType(tt.expectedCondition.Type),
				tt.expectedCondition.Status,
				v1alpha1.ConditionReason(tt.expectedCondition.Reason),
				tt.expectedCondition.Message,
			)
		})
	}
}

func minimalPod(ready bool) *corev1.Pod {
	conditions := []corev1.PodCondition{}
	if ready {
		conditions = append(conditions, corev1.PodCondition{
			Type:    corev1.PodReady,
			Status:  corev1.ConditionTrue,
			Reason:  "ContainersReady",
			Message: "all containers in pod are ready",
		})
	} else {
		conditions = append(conditions, corev1.PodCondition{
			Type:    corev1.PodReady,
			Status:  corev1.ConditionFalse,
			Reason:  "ContainersNotReady",
			Message: "containers with unready status: [a]",
		})
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rp-pod",
			Namespace: "wherever",
			Labels: map[string]string{
				v1alpha1.LabelApp: "connection",
			},
		},
		Status: corev1.PodStatus{
			Conditions: conditions,
			Phase:      corev1.PodRunning,
		},
	}
}
