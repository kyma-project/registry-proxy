package state

import (
	"context"
	"reflect"
	"time"

	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/fsm"
	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/resources"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// sFnHandleService handles creation of the service
func sFnHandleService(ctx context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	service, err := getService(ctx, m)
	if err != nil {
		return nil, nil, err
	}

	if service == nil {
		m.State.Connection.Status.NodePort = 0

		return createService(ctx, m)
	}

	m.State.Service = service
	// Does it match wanted state
	requeueNeeded, err := updateServiceIfNeeded(ctx, m)
	if err != nil {
		return nil, nil, err
	}
	if requeueNeeded {
		return requeueAfter(time.Minute)
	}

	if m.State.Service.Spec.Ports[0].NodePort == 0 {
		// nodePort not ready yet
		return requeueAfter(time.Minute)
	}

	m.State.NodePort = m.State.Service.Spec.Ports[0].NodePort
	return nextState(sFnHandleStatus)
}

func getService(ctx context.Context, m *fsm.StateMachine) (*corev1.Service, error) {
	currentService := &corev1.Service{}
	rp := m.State.Connection

	serviceErr := m.Client.Get(ctx, client.ObjectKey{
		Namespace: rp.GetNamespace(),
		Name:      rp.GetName(),
	}, currentService)
	if serviceErr != nil {
		// Service not existing is expected behavior
		if errors.IsNotFound(serviceErr) {
			return nil, nil
		}
		m.Log.Error(serviceErr, "unable to fetch Service for RegistryProxy")
		return nil, serviceErr
	}

	return currentService, nil
}

func createService(ctx context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	service := resources.NewService(&m.State.Connection)

	// Set the ownerRef for the Service, ensuring that the Service
	// will be deleted when the Function CR is deleted.
	if err := controllerutil.SetControllerReference(&m.State.Connection, service, m.Scheme); err != nil {
		m.Log.Error(err, "failed to set controller reference on Service")
		return stopWithEventualError(err)
	}

	if err := m.Client.Create(ctx, service); err != nil {
		m.Log.Error(err, "failed to create new Service", "Service.Namespace", service.GetNamespace(), "Service.Name", service.GetName())
		return stopWithEventualError(err)
	}

	return requeueAfter(time.Minute)
}

func updateServiceIfNeeded(ctx context.Context, m *fsm.StateMachine) (bool, error) {
	wantedService := resources.NewService(&m.State.Connection)
	if !serviceChanged(m.State.Service, wantedService) {
		return false, nil
	}

	m.State.Service.Spec.Ports = wantedService.Spec.Ports
	m.State.Service.Spec.Selector = wantedService.Spec.Selector
	m.State.Service.Labels = wantedService.Labels
	m.State.Service.Spec.Type = wantedService.Spec.Type

	return updateService(ctx, m)
}

func serviceChanged(got, wanted *corev1.Service) bool {
	gotS := got.Spec
	wantedS := wanted.Spec

	labelsChanged := !reflect.DeepEqual(got.Labels, wanted.Labels)
	portChanged := !reflect.DeepEqual(gotS.Ports[0].Port, wantedS.Ports[0].Port)
	protocolChanged := !reflect.DeepEqual(gotS.Ports[0].Protocol, wantedS.Ports[0].Protocol)
	targetChanged := !reflect.DeepEqual(gotS.Ports[0].TargetPort, wantedS.Ports[0].TargetPort)
	typeChanged := !reflect.DeepEqual(gotS.Type, wantedS.Type)
	selectorChanged := !reflect.DeepEqual(gotS.Selector, wantedS.Selector)

	return labelsChanged ||
		portChanged ||
		protocolChanged ||
		targetChanged ||
		typeChanged ||
		selectorChanged
}

func updateService(ctx context.Context, m *fsm.StateMachine) (bool, error) {
	m.Log.Info("Updating Service %s/%s", m.State.Service.GetNamespace(), m.State.Service.GetName())
	if err := m.Client.Update(ctx, m.State.Service); err != nil {
		m.Log.Error(err, "Failed to update Service", "Service.Namespace", m.State.Service.GetNamespace(), "Service.Name", m.State.Service.GetName())
		return false, err
	}

	// Requeue the request to ensure the Service is updated
	return true, nil
}
