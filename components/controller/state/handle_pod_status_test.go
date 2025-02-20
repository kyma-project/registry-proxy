package state

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.tools.sap/kyma/image-pull-reverse-proxy/components/controller/api/v1alpha1"
	v1alpha2 "github.tools.sap/kyma/image-pull-reverse-proxy/components/controller/api/v1alpha1"
	"github.tools.sap/kyma/image-pull-reverse-proxy/components/controller/fsm"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func probeHandleSuccess(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func probeHandleFailure(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTeapot)
}

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
				ReverseProxy: v1alpha2.ImagePullReverseProxy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "rp",
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
		requireContainsCondition(t, m.State.ReverseProxy.Status, v1alpha1.ConditionRunning, metav1.ConditionFalse, v1alpha1.ConditionReasonProbeError, "no pod exists")
		requireContainsCondition(t, m.State.ReverseProxy.Status, v1alpha1.ConditionReady, metav1.ConditionFalse, v1alpha1.ConditionReasonProbeError, "no pod exists")
	})

	t.Run("one pod exists", func(t *testing.T) {
		successServer := httptest.NewServer(http.HandlerFunc(probeHandleSuccess))
		successURL, err := url.Parse(successServer.URL)
		require.NoError(t, err)
		onePod := minimalPod(successURL)
		onePod.CreationTimestamp = metav1.NewTime(time.Now().Add(-time.Hour))
		scheme := minimalScheme(t)

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(onePod).Build()

		m := fsm.StateMachine{
			State: fsm.SystemState{
				ReverseProxy: v1alpha2.ImagePullReverseProxy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "rp",
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
		require.Nil(t, next)
	})
	t.Run("multiple pods exist", func(t *testing.T) {
		// create two distinctive fake probes to check if we really took the correct pod depending on returned condition
		failureServer := httptest.NewServer(http.HandlerFunc(probeHandleFailure))
		failureURL, err := url.Parse(failureServer.URL)
		require.NoError(t, err)
		successServer := httptest.NewServer(http.HandlerFunc(probeHandleSuccess))
		successURL, err := url.Parse(successServer.URL)
		require.NoError(t, err)

		// create two pods with different creation timestamps, the first has failing probes
		firstPod := minimalPod(failureURL)
		firstPod.CreationTimestamp = metav1.NewTime(time.Now().Add(-time.Hour))
		secondPod := minimalPod(successURL)
		secondPod.Name = "rp-pod2"
		secondPod.CreationTimestamp = metav1.NewTime(time.Now().Add(-time.Minute))
		scheme := minimalScheme(t)

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(firstPod, secondPod).Build()

		m := fsm.StateMachine{
			State: fsm.SystemState{
				ReverseProxy: v1alpha2.ImagePullReverseProxy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "rp",
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
		require.Nil(t, next)
		requireContainsCondition(t, m.State.ReverseProxy.Status, v1alpha1.ConditionRunning, metav1.ConditionTrue, v1alpha1.ConditionReasonProbeSuccess, "")
	})
}

func TestHandleProbe(t *testing.T) {

	failureServer := httptest.NewServer(http.HandlerFunc(probeHandleFailure))
	successServer := httptest.NewServer(http.HandlerFunc(probeHandleSuccess))
	failureURL, err := url.Parse(failureServer.URL)
	require.NoError(t, err)
	successURL, err := url.Parse(successServer.URL)
	require.NoError(t, err)

	tests := []struct {
		name              string
		rp                *v1alpha1.ImagePullReverseProxy
		podIP             string
		probe             *corev1.Probe
		expectedCondition metav1.Condition
	}{
		{
			name:  "should return error on broken probe",
			rp:    &v1alpha1.ImagePullReverseProxy{},
			podIP: "127.0.0.1",
			probe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Port: intstr.FromInt(8080),
						Path: "/healthz",
					},
				},
			},
			expectedCondition: metav1.Condition{
				Type:    string(v1alpha1.ConditionRunning),
				Status:  metav1.ConditionFalse,
				Reason:  string(v1alpha1.ConditionReasonProbeError),
				Message: "cannot read health probe:Get \"http://127.0.0.1:8080/healthz\": dial tcp 127.0.0.1:8080: connect: connection refused",
			},
		},
		{
			name:  "should return probe failure",
			rp:    &v1alpha1.ImagePullReverseProxy{},
			podIP: failureURL.Hostname(),
			probe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Port: intstr.FromString(failureURL.Port()),
						Path: "/healthz",
					},
				},
			},
			expectedCondition: metav1.Condition{
				Type:    string(v1alpha1.ConditionRunning),
				Status:  metav1.ConditionFalse,
				Reason:  string(v1alpha1.ConditionReasonProbeFailure),
				Message: "/healthz probe has returned 418 status",
			},
		},
		{
			name:  "should return probe success",
			rp:    &v1alpha1.ImagePullReverseProxy{},
			podIP: successURL.Hostname(),
			probe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Port: intstr.FromString(successURL.Port()),
						Path: "/healthz",
					},
				},
			},
			expectedCondition: metav1.Condition{
				Type:    string(v1alpha1.ConditionRunning),
				Status:  metav1.ConditionTrue,
				Reason:  string(v1alpha1.ConditionReasonProbeSuccess),
				Message: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handleProbe(tt.rp, tt.podIP, tt.probe, v1alpha1.ConditionType(tt.expectedCondition.Type))
			requireContainsCondition(t, tt.rp.Status,
				v1alpha1.ConditionType(tt.expectedCondition.Type),
				tt.expectedCondition.Status,
				v1alpha1.ConditionReason(tt.expectedCondition.Reason),
				tt.expectedCondition.Message,
			)
		})
	}
}

func minimalPod(probesURL *url.URL) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rp-pod",
			Namespace: "wherever",
			Labels: map[string]string{
				"app": "rp",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					LivenessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Port: intstr.FromString(probesURL.Port()),
								Path: "/probe",
							},
						},
					},
					ReadinessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Port: intstr.FromString(probesURL.Port()),
								Path: "/probe",
							},
						},
					},
				},
			},
		},
		Status: corev1.PodStatus{
			PodIP: probesURL.Hostname(),
		},
	}
}
