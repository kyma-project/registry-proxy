package state

import (
	"context"
	"fmt"
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
	err = handleLivenessStatus(&m.State.Connection, pod.Status.Phase)
	if err != nil {
		return stopWithEventualError(err)
	}
	err = handleReadinessStatus(&m.State.Connection, pod.Status.Conditions)
	if err != nil {
		return stopWithEventualError(err)
	}

	return nextState(sFnHandleService)
}

func handleLivenessStatus(rp *v1alpha1.Connection, phase corev1.PodPhase) error {
	if phase == corev1.PodRunning {
		rp.UpdateCondition(
			v1alpha1.ConditionConnectionDeployed,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonResourcesDeployed,
			"Reverse-proxy ready",
		)
		return nil
	}
	//nolint: staticcheck
	err := fmt.Errorf("Reverse-proxy not ready: pod is in phase %s", phase)
	rp.UpdateCondition(
		v1alpha1.ConditionConnectionDeployed,
		metav1.ConditionFalse,
		v1alpha1.ConditionReasonError,
		err.Error(),
	)
	return err
}

func handleReadinessStatus(rp *v1alpha1.Connection, conditions []corev1.PodCondition) error {
	condition := getCondition(conditions, corev1.PodReady)
	if condition == nil {
		//nolint: staticcheck
		err := fmt.Errorf("Target registry not reachable: no condition found")
		rp.UpdateCondition(
			v1alpha1.ConditionConnectionReady,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonNotEstabilished,
			err.Error(),
		)
		return err
	}
	if condition.Status == corev1.ConditionTrue {
		rp.UpdateCondition(
			v1alpha1.ConditionConnectionReady,
			metav1.ConditionTrue,
			v1alpha1.ConditionReasonEstabilished,
			"Target registry reachable",
		)
		return nil
	}
	//nolint: staticcheck
	err := fmt.Errorf("Target registry not reachable: %s", condition.Reason)
	rp.UpdateCondition(
		v1alpha1.ConditionConnectionReady,
		metav1.ConditionFalse,
		v1alpha1.ConditionReasonNotEstabilished,
		err.Error(),
	)
	return err
}

func getCondition(conditions []corev1.PodCondition, conditionType corev1.PodConditionType) *corev1.PodCondition {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return &condition
		}
	}
	return nil
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
