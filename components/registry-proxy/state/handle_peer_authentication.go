package state

import (
	"context"
	"os"
	"reflect"
	"time"

	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/fsm"
	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/resources"
	securityclientv1 "istio.io/client-go/pkg/apis/security/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func sFnHandlePeerAuthentication(ctx context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	nextFn := sFnHandleStatus
	if os.Getenv("ISTIO_INSTALLED") != "true" {
		return nextState(nextFn)
	}
	pa, err := getPeerAuthentication(ctx, m)
	if err != nil {
		return nil, nil, err
	}

	if pa == nil {
		return createPeerAuthentication(ctx, m)
	}

	m.State.PeerAuthentication = pa

	// Does it match wanted state
	requeueNeeded, err := updatePeerAuthenticationIfNeeded(ctx, m)
	if err != nil {
		return nil, nil, err
	}
	if requeueNeeded {
		return requeueAfter(time.Minute)
	}

	return nextState(nextFn)
}

func getPeerAuthentication(ctx context.Context, m *fsm.StateMachine) (*securityclientv1.PeerAuthentication, error) {
	currentPeerAuthentication := &securityclientv1.PeerAuthentication{}
	c := m.State.Connection

	peerAuthenticationErr := m.Client.Get(ctx, client.ObjectKey{
		Namespace: c.GetNamespace(),
		Name:      c.GetName(),
	}, currentPeerAuthentication)

	if peerAuthenticationErr != nil {
		if errors.IsNotFound(peerAuthenticationErr) {
			return nil, nil
		}
		m.Log.Error(peerAuthenticationErr, "unable to fetch PeerAuthentication for RegistryProxy")
		return nil, peerAuthenticationErr
	}
	return currentPeerAuthentication, nil
}

func createPeerAuthentication(ctx context.Context, m *fsm.StateMachine) (fsm.StateFn, *ctrl.Result, error) {
	pa := resources.NewPeerAuthentication(&m.State.Connection)

	// Set the ownerRef for the PeerAuthentication, ensuring that the PeerAuthentication
	// will be deleted when the Function CR is deleted.
	if err := controllerutil.SetControllerReference(&m.State.Connection, pa, m.Scheme); err != nil {
		m.Log.Error(err, "failed to set controller reference on PeerAuthentication")
		return stopWithEventualError(err)
	}

	if err := m.Client.Create(ctx, pa); err != nil {
		m.Log.Error(err, "failed to create new PeerAuthentication", "PeerAuthentication.Namespace", pa.GetNamespace(), "PeerAuthentication.Name", pa.GetName())
		return stopWithEventualError(err)
	}

	return requeueAfter(time.Minute)
}

func updatePeerAuthenticationIfNeeded(ctx context.Context, m *fsm.StateMachine) (bool, error) {
	wantedPA := resources.NewPeerAuthentication(&m.State.Connection)
	if !peerAuthenticationChanged(m.State.PeerAuthentication, wantedPA) {
		return false, nil
	}

	m.State.PeerAuthentication.Spec.Mtls = wantedPA.Spec.Mtls
	m.State.PeerAuthentication.Spec.Selector = wantedPA.Spec.Selector

	return updatePeerAuthentication(ctx, m)
}

func peerAuthenticationChanged(got, wanted *securityclientv1.PeerAuthentication) bool {
	gotS := &got.Spec
	wantedS := &wanted.Spec

	mtlsChanged := !reflect.DeepEqual(gotS.Mtls, wantedS.Mtls)
	selectorChanged := !reflect.DeepEqual(gotS.Selector, wantedS.Selector)

	return mtlsChanged || selectorChanged
}

func updatePeerAuthentication(ctx context.Context, m *fsm.StateMachine) (bool, error) {
	m.Log.Info("Updating PeerAuthentication %s/%s", m.State.PeerAuthentication.GetNamespace(), m.State.PeerAuthentication.GetName())
	if err := m.Client.Update(ctx, m.State.PeerAuthentication); err != nil {
		m.Log.Error(err, "Failed to update PeerAuthentication", "PeerAuthentication.Namespace", m.State.PeerAuthentication.GetNamespace(), "PeerAuthentication.Name", m.State.PeerAuthentication.GetName())
		return false, err
	}

	// Requeue the request to ensure the PeerAuthentication is updated
	return true, nil
}
