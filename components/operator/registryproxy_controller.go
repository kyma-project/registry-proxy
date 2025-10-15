package operator

import (
	"context"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kyma-project/registry-proxy/components/common/cache"
	"github.com/kyma-project/registry-proxy/components/operator/api/v1alpha1"
	"github.com/kyma-project/registry-proxy/components/operator/chart"
	"github.com/kyma-project/registry-proxy/components/operator/fsm"
	"github.com/kyma-project/registry-proxy/components/operator/state"
)

// RegistryProxyReconciler reconciles a RegistryProxy object
type RegistryProxyReconciler struct {
	client.Client
	*rest.Config
	Scheme                     *runtime.Scheme
	Log                        *zap.SugaredLogger
	ConnectivityProxyReadiness cache.BoolCache
	IstioReadiness             cache.BoolCache
	ChartCache                 chart.ManifestCache
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

	sm := fsm.New(r.Client, r.Config, &registryProxy, state.StartState(), r.Scheme, log, r.ConnectivityProxyReadiness, r.IstioReadiness, r.ChartCache)
	return sm.Reconcile(ctx)
}

func (r *RegistryProxyReconciler) retriggerAllRegistryProxyCRs(ctx context.Context, e event.DeleteEvent, q workqueue.TypedRateLimitingInterface[ctrl.Request]) {
	log := r.Log.With("deletion_watcher")

	list := &v1alpha1.RegistryProxyList{}
	err := r.List(ctx, list, &client.ListOptions{})
	if err != nil {
		log.Errorf("error listing registryproxy objects: %s", err.Error())
		return
	}

	for _, s := range list.Items {
		log.Debugf("retriggering reconciliation for RegistryProxy %s/%s", s.GetNamespace(), s.GetName())
		q.Add(ctrl.Request{NamespacedName: client.ObjectKey{
			Namespace: s.GetNamespace(),
			Name:      s.GetName(),
		}})
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *RegistryProxyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	connectivityProxyWatcher := &ConnectivityProxyReadinessWatcher{
		Cache:  r.ConnectivityProxyReadiness,
		Client: r.Client,
		Log:    r.Log,
	}
	istioWatcher := &IstioReadinessWatcher{
		Client: r.Client,
		Cache:  r.IstioReadiness,
		Log:    r.Log,
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.RegistryProxy{}).
		Watches(&v1alpha1.RegistryProxy{}, &handler.Funcs{
			// retrigger all RegistryProxy CRs reconciliations when one is deleted
			// this should ensure at least one RegistryProxy CR is served
			DeleteFunc: r.retriggerAllRegistryProxyCRs,
		}).
		Watches(
			&appsv1.StatefulSet{},
			handler.EnqueueRequestsFromMapFunc(connectivityProxyWatcher.triggerRegistryProxyRequeueOnChange),
			builder.WithPredicates(connectivityProxyWatcher.buildPredicate()),
		).
		Watches(
			&appsv1.Deployment{},
			handler.EnqueueRequestsFromMapFunc(istioWatcher.triggerRegistryProxyRequeueOnChange),
			builder.WithPredicates(istioWatcher.buildPredicate()),
		).
		Named("registry-proxy").
		Complete(r)
}

// ConnectivityProxyReadinessWatcher reconciles all RegistryProxy objects when ConnectivityProxy module is changed
type ConnectivityProxyReadinessWatcher struct {
	client.Client
	Log   *zap.SugaredLogger
	Cache cache.BoolCache
}

func (w *ConnectivityProxyReadinessWatcher) buildPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return w.isConnectivityProxyStatefulSet(e.Object)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return w.isConnectivityProxyStatefulSet(e.ObjectNew)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return w.isConnectivityProxyStatefulSet(e.Object)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return w.isConnectivityProxyStatefulSet(e.Object)
		},
	}
}

func (w *ConnectivityProxyReadinessWatcher) triggerRegistryProxyRequeueOnChange(ctx context.Context, obj client.Object) []reconcile.Request {
	statefulSet, ok := obj.(*appsv1.StatefulSet)
	if !ok {
		w.Log.Errorf("object %s/%s is not a StatefulSet", obj.GetNamespace(), obj.GetName())
		return nil
	}

	// has available replicas and is not being deleted
	isCPReady := statefulSet.Status.AvailableReplicas > 0 && statefulSet.GetDeletionTimestamp() == nil

	if isCPReady == w.Cache.Get() {
		w.Log.Debugf("readiness state of Connectivity Proxy StatefulSet %s/%s has not changed, skipping requeue",
			statefulSet.GetNamespace(), statefulSet.GetName())
		return nil
	}

	w.Log.Infof("Connectivity Proxy readiness changed to: %t, retriggering all Connection CRs' reconciliation", isCPReady)

	list := &v1alpha1.RegistryProxyList{}
	err := w.List(ctx, list)
	if err != nil {
		w.Log.Errorf("failed to list RegistryProxy objects: %v", err)
		return nil
	}

	requests, err := getRegistryProxyReconcilationList(ctx, w.Client)
	if err != nil {
		w.Log.Errorf("failed to get RegistryProxy reconciliation list: %v", err)
		return nil
	}

	w.Cache.Set(isCPReady)
	return requests
}

func (w *ConnectivityProxyReadinessWatcher) isConnectivityProxyStatefulSet(obj client.Object) bool {
	return obj.GetName() == "connectivity-proxy" && obj.GetNamespace() == "kyma-system"
}

type IstioReadinessWatcher struct {
	client.Client
	Log   *zap.SugaredLogger
	Cache cache.BoolCache
}

func (w *IstioReadinessWatcher) buildPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return w.isIstioDeployment(e.Object)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return w.isIstioDeployment(e.ObjectNew)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return w.isIstioDeployment(e.Object)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return w.isIstioDeployment(e.Object)
		},
	}
}

func (w *IstioReadinessWatcher) triggerRegistryProxyRequeueOnChange(ctx context.Context, obj client.Object) []reconcile.Request {
	deployment, ok := obj.(*appsv1.Deployment)
	if !ok {
		w.Log.Errorf("object %s/%s is not a Deployment", obj.GetNamespace(), obj.GetName())
		return nil
	}

	// has available replicas and is not being deleted
	isIstioReady := deployment.Status.AvailableReplicas > 0 && deployment.GetDeletionTimestamp() == nil

	if isIstioReady == w.Cache.Get() {
		w.Log.Debugf("readiness state of Istio Deployment %s/%s has not changed, skipping requeue",
			deployment.GetNamespace(), deployment.GetName())
		return nil
	}

	w.Log.Infof("Connectivity Proxy readiness changed to: %t, retriggering all Connection CRs' reconciliation", isIstioReady)

	list := &v1alpha1.RegistryProxyList{}
	err := w.List(ctx, list)
	if err != nil {
		w.Log.Errorf("failed to list RegistryProxy objects: %v", err)
		return nil
	}

	requests, err := getRegistryProxyReconcilationList(ctx, w.Client)
	if err != nil {
		w.Log.Errorf("failed to get RegistryProxy reconciliation list: %v", err)
		return nil
	}

	w.Cache.Set(isIstioReady)
	return requests
}

func getRegistryProxyReconcilationList(ctx context.Context, c client.Client) ([]reconcile.Request, error) {
	list := &v1alpha1.RegistryProxyList{}
	err := c.List(ctx, list)
	if err != nil {
		return nil, err
	}

	requests := []reconcile.Request{}
	for _, rp := range list.Items {
		// collect all RegistryProxy objects
		requests = append(requests, reconcile.Request{
			NamespacedName: client.ObjectKey{
				Name:      rp.GetName(),
				Namespace: rp.GetNamespace(),
			},
		})
	}
	return requests, nil
}

func (w *IstioReadinessWatcher) isIstioDeployment(obj client.Object) bool {
	return obj.GetName() == "istiod" && obj.GetNamespace() == "istio-system"
}
