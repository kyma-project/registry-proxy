package controller

import (
	"context"
	"time"

	"github.tools.sap/kyma/registry-proxy/components/common/cache"
	"go.uber.org/zap"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// TODO: rename this to something more meaningful. ConnectivityProxyReconciler
// TODO: move this to the operator/common?
type CrdsReconciler struct {
	client.Client
	Log   *zap.SugaredLogger
	Cache cache.BoolCache
}

// SetupWithManager sets up the controller with the Manager.
func (r *CrdsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiextensionsv1.CustomResourceDefinition{}).
		WithEventFilter(r.buildPredicates()).
		Named("rp-crds-controller").
		Complete(r)
}

func (r *CrdsReconciler) buildPredicates() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			obj, ok := e.Object.(*apiextensionsv1.CustomResourceDefinition)
			if !ok {
				return false
			}
			return r.isConnectivityProxy(obj)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			obj, ok := e.ObjectNew.(*apiextensionsv1.CustomResourceDefinition)
			if !ok {
				return false
			}
			return r.isConnectivityProxy(obj)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			obj, ok := e.Object.(*apiextensionsv1.CustomResourceDefinition)
			if !ok {
				return false
			}
			return r.isConnectivityProxy(obj)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			obj, ok := e.Object.(*apiextensionsv1.CustomResourceDefinition)
			if !ok {
				return false
			}
			return r.isConnectivityProxy(obj)
		},
	}
}

func (r *CrdsReconciler) isConnectivityProxy(crd *apiextensionsv1.CustomResourceDefinition) bool {
	return crd.Spec.Group == "connectivityproxy.sap.com" && crd.Spec.Names.Kind == "ConnectivityProxy"
}

// Reconcile reads that state of the cluster for a CRD object and makes changes based on the state read
func (r *CrdsReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	crd := apiextensionsv1.CustomResourceDefinition{}
	err := r.Client.Get(ctx, request.NamespacedName, &crd)

	if err != nil {
		r.Cache.Set(false)
		r.Log.Errorf("Error getting Connectivity Proxy CRD: %s\n", err)
		return ctrl.Result{}, nil
	}
	r.Cache.Set(true)
	return ctrl.Result{RequeueAfter: time.Minute}, nil
}
