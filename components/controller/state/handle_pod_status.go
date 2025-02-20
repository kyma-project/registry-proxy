package state

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.tools.sap/kyma/image-pull-reverse-proxy/components/controller/api/v1alpha1"
	"github.tools.sap/kyma/image-pull-reverse-proxy/components/controller/fsm"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// sFnHandlePodStatus checks healthz/readyz probes of the latest pod and updates the conditions in the CR
func sFnHandlePodStatus(ctx context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	podList := &corev1.PodList{}
	matchLabels := client.MatchingLabels{}
	matchLabels["app"] = m.State.ReverseProxy.Name
	// TODO: do we have to set up cache to list only our Pods
	err := m.Client.List(ctx, podList, matchLabels)
	if err != nil {
		return nil, nil, err
	}
	// check pod's healthz and readyz
	if len(podList.Items) < 1 {
		// no pod exists, reset conditions and retry
		m.State.ReverseProxy.UpdateCondition(
			v1alpha1.ConditionRunning,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonProbeError,
			"no pod exists",
		)
		m.State.ReverseProxy.UpdateCondition(
			v1alpha1.ConditionReady,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonProbeError,
			"no pod exists",
		)
		return requeueAfter(time.Minute)
	}

	pod := getLatestPod(podList)
	if pod.Status.PodIP == "" {
		// podIP is not instantly set, sometimes we have to wait for it
		return requeueAfter(time.Second * 10)
	}
	err = handleProbe(&m.State.ReverseProxy, pod.Status.PodIP, pod.Spec.Containers[0].LivenessProbe, v1alpha1.ConditionRunning)
	if err != nil {
		return stopWithEventualError(err)
	}
	err = handleProbe(&m.State.ReverseProxy, pod.Status.PodIP, pod.Spec.Containers[0].ReadinessProbe, v1alpha1.ConditionReady)
	if err != nil {
		return stopWithEventualError(err)
	}

	// TODO: next function state
	return nextState(nil)
}

// TODO: return status, we should requeue on error I guess
func handleProbe(rp *v1alpha1.ImagePullReverseProxy, podIP string, probe *corev1.Probe, condition v1alpha1.ConditionType) error {
	probeStatus, err := getProbeStatus(podIP, probe)
	if err != nil {
		rp.UpdateCondition(
			condition,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonProbeError,
			fmt.Sprintf("cannot read health probe:%s", err),
		)
		return err
	} else if probeSuccessful(probeStatus) {
		rp.UpdateCondition(
			condition,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonProbeSuccess,
			"",
		)
	} else {
		rp.UpdateCondition(
			condition,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonProbeFailure,
			fmt.Sprintf("%s probe has returned %d status", probe.HTTPGet.Path, probeStatus),
		)
	}
	return nil
}

// getProbeStatus checks status of a HTTPGet probe
func getProbeStatus(podIP string, probe *corev1.Probe) (int, error) {
	probeURL := fmt.Sprintf("http://%s:%s%s", podIP, probe.HTTPGet.Port.String(), probe.HTTPGet.Path)
	res, err := http.Get(probeURL)
	if err != nil {
		return 0, err
	}
	return res.StatusCode, nil
}

func getLatestPod(podList *corev1.PodList) *corev1.Pod {
	if len(podList.Items) < 1 {
		return nil
	}

	latestPod := 0
	lastCreationTimestamp := podList.Items[0].CreationTimestamp
	for i, pod := range podList.Items {
		if lastCreationTimestamp.Before(&pod.CreationTimestamp) {
			latestPod = i
			lastCreationTimestamp = pod.CreationTimestamp
		}
	}
	return &podList.Items[latestPod]
}

func probeSuccessful(status int) bool {
	return status >= 200 && status < 300
}
