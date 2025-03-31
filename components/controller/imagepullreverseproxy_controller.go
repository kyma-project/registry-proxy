package controller

import (
	"context"
	"github.tools.sap/kyma/image-pull-reverse-proxy/components/controller/api/v1alpha1"
	"github.tools.sap/kyma/image-pull-reverse-proxy/components/controller/cache"
	"github.tools.sap/kyma/image-pull-reverse-proxy/components/controller/fsm"
	"github.tools.sap/kyma/image-pull-reverse-proxy/components/controller/state"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// ImagePullReverseProxyReconciler reconciles a ImagePullReverseProxy object
type ImagePullReverseProxyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    *zap.SugaredLogger
	Cache  cache.BoolCache
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/reconcile
func (r *ImagePullReverseProxyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.With("request", req)
	log.Info("reconciliation started")

	var reverseProxy v1alpha1.ImagePullReverseProxy
	if err := r.Get(ctx, req.NamespacedName, &reverseProxy); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	sm := fsm.New(r.Client, &reverseProxy, state.StartState(), r.Scheme, log, r.Cache)
	return sm.Reconcile(ctx)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ImagePullReverseProxyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.ImagePullReverseProxy{}).
		WithEventFilter(buildPredicates()).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Pod{}).
		Named("image-pull-reverse-proxy").
		Complete(r)
}

func buildPredicates() predicate.Funcs {
	// Predicate to skip reconciliation when the object is being deleted
	return predicate.Funcs{
		// Allow create events
		UpdateFunc: func(e event.UpdateEvent) bool {
			return true
		},
		// Allow create events
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		// Don't allow delete events
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		// Allow generic events (e.g., external triggers)
		GenericFunc: func(e event.GenericEvent) bool {
			return true
		},
	}
}
