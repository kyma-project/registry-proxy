package operator

import (
	"context"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.tools.sap/kyma/registry-proxy/components/common/cache"
	"github.tools.sap/kyma/registry-proxy/components/operator/api/v1alpha1"
	"github.tools.sap/kyma/registry-proxy/components/operator/chart"
	"github.tools.sap/kyma/registry-proxy/components/operator/fsm"
	"github.tools.sap/kyma/registry-proxy/components/operator/state"
)

// RegistryProxyReconciler reconciles a RegistryProxy object
type RegistryProxyReconciler struct {
	client.Client
	*rest.Config
	Scheme     *runtime.Scheme
	Log        *zap.SugaredLogger
	Cache      cache.BoolCache
	ChartCache chart.ManifestCache
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *RegistryProxyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.With("request", req)
	log.Info("reconciliation started")

	var registryProxy v1alpha1.RegistryProxy
	if err := r.Get(ctx, req.NamespacedName, &registryProxy); err != nil {
		log.Error(err, "unable to fetch RegistryProxy")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	sm := fsm.New(r.Client, r.Config, &registryProxy, state.StartState(), r.Scheme, log, r.Cache, r.ChartCache)
	return sm.Reconcile(ctx)
}

// SetupWithManager sets up the controller with the Manager.
func (r *RegistryProxyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.RegistryProxy{}).
		Named("registry-proxy").
		Complete(r)
}
