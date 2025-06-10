package state

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/api/v1alpha1"
	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/fsm"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// sFnHandlePodStatus checks healthz/readyz probes of the latest pod and updates the conditions in the CR
func sFnHandlePodStatus(ctx context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	podList := &corev1.PodList{}
	matchLabels := client.MatchingLabels{}
	matchLabels[v1alpha1.LabelApp] = m.State.Connection.Name
	err := m.Client.List(ctx, podList, matchLabels)
	if err != nil {
		return nil, nil, err
	}
	// check pod's healthz and readyz
	if len(podList.Items) < 1 {
		// no pod exists, reset conditions and retry
		m.State.Connection.UpdateCondition(
			v1alpha1.ConditionConnectionDeployed,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonResourcesNotReady,
			"no pod exists",
		)
		m.State.Connection.UpdateCondition(
			v1alpha1.ConditionConnectionReady,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonNotEstabilished,
			"no pod exists",
		)
		return requeueAfter(time.Minute)
	}

	pod := GetLatestPod(podList)
	if pod.Status.PodIP == "" {
		// podIP is not instantly set, sometimes we have to wait for it
		return requeueAfter(time.Second * 10)
	}
	err = handleLivenessProbe(&m.State.Connection, pod.Status.PodIP, pod.Spec.Containers[0].LivenessProbe)
	if err != nil {
		return stopWithEventualError(err)
	}
	err = handleReadinessProbe(&m.State.Connection, pod.Status.PodIP, pod.Spec.Containers[0].ReadinessProbe)
	if err != nil {
		return stopWithEventualError(err)
	}

	return nextState(sFnHandleService)
}

func handleLivenessProbe(rp *v1alpha1.Connection, podIP string, probe *corev1.Probe) error {
	probeStatus, err := getProbeStatus(podIP, probe)
	if err != nil {
		rp.UpdateCondition(
			v1alpha1.ConditionConnectionDeployed,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonError,
			fmt.Sprintf("Reverse-proxy not ready: %s", err),
		)
		return err
	} else if probeSuccessful(probeStatus) {
		rp.UpdateCondition(
			v1alpha1.ConditionConnectionDeployed,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonResourcesDeployed,
			"Reverse-proxy ready",
		)
		return nil
	}
	rp.UpdateCondition(
		v1alpha1.ConditionConnectionDeployed,
		metav1.ConditionFalse,
		v1alpha1.ConditionReasonResourcesNotReady,
		fmt.Sprintf("Reverse-proxy not ready: probe has returned status %d", probeStatus),
	)
	return nil
}

func handleReadinessProbe(rp *v1alpha1.Connection, podIP string, probe *corev1.Probe) error {
	probeStatus, err := getProbeStatus(podIP, probe)
	if err != nil {
		rp.UpdateCondition(
			v1alpha1.ConditionConnectionReady,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonError,
			fmt.Sprintf("Target registry not reachable: %s", err),
		)
		return err
	} else if probeSuccessful(probeStatus) {
		rp.UpdateCondition(
			v1alpha1.ConditionConnectionReady,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonEstabilished,
			"Target registry reachable",
		)
		return nil
	}
	rp.UpdateCondition(
		v1alpha1.ConditionConnectionReady,
		metav1.ConditionFalse,
		v1alpha1.ConditionReasonNotEstabilished,
		fmt.Sprintf("Target registry not reachable: probe has returned status %d", probeStatus),
	)
	return nil
}

// getProbeStatus checks status of a HTTPGet probe
func getProbeStatus(podIP string, probe *corev1.Probe) (int, error) {
	if probe == nil || probe.HTTPGet == nil {
		return 0, fmt.Errorf("probe is nil")
	}
	probeURL := fmt.Sprintf("http://%s:%s%s", podIP, probe.HTTPGet.Port.String(), probe.HTTPGet.Path)
	res, err := http.Get(probeURL)
	if err != nil {
		return 0, err
	}
	return res.StatusCode, nil
}

func GetLatestPod(podList *corev1.PodList) *corev1.Pod {
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
