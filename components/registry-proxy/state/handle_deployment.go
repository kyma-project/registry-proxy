package state

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.tools.sap/kyma/registry-proxy/components/common/container"
	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/api/v1alpha1"
	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/fsm"
	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/resources"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// sFnHandleDeployment is responsible for handling the deployment
func sFnHandleDeployment(ctx context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	// #1 Does it exist
	deployment, err := getDeployment(ctx, m)
	if err != nil {
		return nil, nil, err
	}
	// Create if not
	if deployment == nil {
		return createDeployment(ctx, m)
	}

	m.State.Deployment = deployment
	// #2 Does it match CR
	requeueNeeded, err := updateDeploymentIfNeeded(ctx, m)
	if err != nil {
		return nil, nil, err
	}
	if requeueNeeded {
		return requeueAfter(time.Minute)
	}

	// Move on to next state if all are true
	// If not Create/Update or Requeue and remove historical info about the pod
	return nextState(sFnHandlePodStatus)
}

func getDeployment(ctx context.Context, m *fsm.StateMachine) (*appsv1.Deployment, error) {
	currentDeployment := &appsv1.Deployment{}
	rp := m.State.Connection
	deploymentErr := m.Client.Get(ctx, client.ObjectKey{
		Namespace: rp.GetNamespace(),
		Name:      rp.GetName(),
	}, currentDeployment)
	if deploymentErr != nil {
		if errors.IsNotFound(deploymentErr) { // Deployment not existing is expected behavior
			return nil, nil
		}
		m.Log.Error(deploymentErr, "unable to fetch Deployment for Connection")
		return nil, deploymentErr
	}

	return currentDeployment, nil
}

func createDeployment(ctx context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	deployment := resources.NewDeployment(&m.State.Connection, m.State.ProxyURL, m.State.AuthorizationNodePort)

	// Set the ownerRef for the Deployment, ensuring that the Deployment
	// will be deleted when the RP CR is deleted.
	if err := controllerutil.SetControllerReference(&m.State.Connection, deployment, m.Scheme); err != nil {
		m.Log.Error(err, "failed to set controller reference on Deployment")
		m.State.Connection.UpdateCondition( // We update the condition on every possible return to make sure it's up to date
			v1alpha1.ConditionConnectionDeployed,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonDeploymentFailed,
			"failed to set controller reference on Deployment")
		return stopWithEventualError(err)
	}

	if err := m.Client.Create(ctx, deployment); err != nil {
		m.Log.Error(err, "failed to create new Deployment", "Deployment.Namespace", deployment.GetNamespace(), "Deployment.Name", deployment.GetName())
		m.State.Connection.UpdateCondition(
			v1alpha1.ConditionConnectionDeployed,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonDeploymentFailed,
			fmt.Sprintf("Deployment %s create failed: %s", deployment.GetName(), err.Error()),
		)
		return stopWithEventualError(err)
	}
	m.State.Connection.UpdateCondition(
		v1alpha1.ConditionConnectionDeployed,
		metav1.ConditionUnknown,
		v1alpha1.ConditionReasonDeploymentCreated,
		fmt.Sprintf("Deployment %s created", deployment.GetName()),
	)

	return requeueAfter(time.Minute)
}

func updateDeploymentIfNeeded(ctx context.Context, m *fsm.StateMachine) (bool, error) {
	wantedDeployment := resources.NewDeployment(&m.State.Connection, m.State.ProxyURL, m.State.AuthorizationNodePort)
	if !deploymentChanged(m.State.Deployment, wantedDeployment) {
		return false, nil
	}

	m.State.Deployment.Spec.Template = wantedDeployment.Spec.Template
	m.State.Deployment.Spec.Replicas = wantedDeployment.Spec.Replicas
	m.State.Deployment.Spec.Selector = wantedDeployment.Spec.Selector
	return updateDeployment(ctx, m)
}

func deploymentChanged(got, wanted *appsv1.Deployment) bool {
	if len(got.Spec.Template.Spec.Containers) < 1 ||
		len(wanted.Spec.Template.Spec.Containers) < 1 ||
		len(got.Spec.Template.Spec.Containers) != len(wanted.Spec.Template.Spec.Containers) {
		return true
	}

	gotRegistryC := container.Get(got.Spec.Template.Spec.Containers, resources.RegistryContainerName)
	wantedRegistryC := container.Get(wanted.Spec.Template.Spec.Containers, resources.RegistryContainerName)
	if gotRegistryC == nil || wantedRegistryC == nil {
		return true
	}
	registryContainerChanged := containerChanged(*gotRegistryC, *wantedRegistryC)

	authorizationContainerChanged := false
	wantedAuthorizationC := container.Get(wanted.Spec.Template.Spec.Containers, resources.AuthorizationContainerName)
	if wantedAuthorizationC != nil {
		gotAuthorizationC := container.Get(got.Spec.Template.Spec.Containers, resources.AuthorizationContainerName)
		if gotAuthorizationC == nil {
			return true
		}
		authorizationContainerChanged = containerChanged(*gotAuthorizationC, *wantedAuthorizationC)
	}

	labelsChanged := !reflect.DeepEqual(got.Spec.Template.Labels, wanted.Spec.Template.Labels)

	replicasChanged := (got.Spec.Replicas == nil && wanted.Spec.Replicas != nil) ||
		(got.Spec.Replicas != nil && wanted.Spec.Replicas == nil) ||
		(got.Spec.Replicas != nil && wanted.Spec.Replicas != nil && *got.Spec.Replicas != *wanted.Spec.Replicas)

	return registryContainerChanged ||
		authorizationContainerChanged ||
		labelsChanged ||
		replicasChanged
}

func containerChanged(gotC, wantedC corev1.Container) bool {
	imageChanged := gotC.Image != wantedC.Image
	commandChanged := !reflect.DeepEqual(gotC.Command, wantedC.Command)
	resourcesChanged := !reflect.DeepEqual(gotC.Resources, wantedC.Resources)
	argsChanged := !reflect.DeepEqual(gotC.Args, wantedC.Args)
	envChanged := !reflect.DeepEqual(gotC.Env, wantedC.Env)
	portsChanged := !reflect.DeepEqual(gotC.Ports, wantedC.Ports)
	return imageChanged ||
		commandChanged ||
		resourcesChanged ||
		argsChanged ||
		envChanged ||
		portsChanged
}

func updateDeployment(ctx context.Context, m *fsm.StateMachine) (bool, error) {
	m.Log.Infof("Updating Deployment %s/%s", m.State.Deployment.GetNamespace(), m.State.Deployment.GetName())
	if err := m.Client.Update(ctx, m.State.Deployment); err != nil {
		m.Log.Error(err, "Failed to update Deployment", "Deployment.Namespace", m.State.Deployment.GetNamespace(), "Deployment.Name", m.State.Deployment.GetName())
		m.State.Connection.UpdateCondition(
			v1alpha1.ConditionConnectionDeployed,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonDeploymentFailed,
			fmt.Sprintf("Deployment %s update failed: %s", m.State.Deployment.GetName(), err.Error()))
		return false, err
	}
	m.State.Connection.UpdateCondition(
		v1alpha1.ConditionConnectionDeployed,
		metav1.ConditionUnknown,
		v1alpha1.ConditionReasonDeploymentUpdated,
		fmt.Sprintf("Deployment %s updated", m.State.Deployment.GetName()))
	// Requeue the request to ensure the Deployment is updated
	return true, nil
}
