package fsm

import (
	"context"
	"fmt"
	"github.tools.sap/kyma/image-pull-reverse-proxy/components/controller/cache"
	"reflect"
	"runtime"
	"strings"

	"github.tools.sap/kyma/image-pull-reverse-proxy/components/controller/api/v1alpha1"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apimachineryruntime "k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StateFn func(context.Context, *StateMachine) (StateFn, *ctrl.Result, error)

type SystemState struct {
	ReverseProxy   v1alpha1.ImagePullReverseProxy
	statusSnapshot v1alpha1.ImagePullReverseProxyStatus
	ProxyURL       string
	NodePort       int32
	Deployment     *appsv1.Deployment
	Service        *corev1.Service
}

func (s *SystemState) saveStatusSnapshot() {
	result := s.ReverseProxy.Status.DeepCopy()
	if result == nil {
		result = &v1alpha1.ImagePullReverseProxyStatus{}
	}
	s.statusSnapshot = *result
}

type StateMachine struct {
	nextFn StateFn
	State  SystemState
	Log    *zap.SugaredLogger
	Client client.Client
	Scheme *apimachineryruntime.Scheme
	Cache  cache.BoolCache
}

func (m *StateMachine) stateFnName() string {
	fullName := runtime.FuncForPC(reflect.ValueOf(m.nextFn).Pointer()).Name()
	splitFullName := strings.Split(fullName, ".")

	if len(splitFullName) < 3 {
		return fullName
	}

	shortName := splitFullName[len(splitFullName)-1]
	return shortName
}

func (m *StateMachine) Reconcile(ctx context.Context) (ctrl.Result, error) {
	var err error
	var result *ctrl.Result
loop:
	for m.nextFn != nil && err == nil {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			break loop

		default:
			m.Log.Info(fmt.Sprintf("switching state: %s", m.stateFnName()))
			m.nextFn, result, err = m.nextFn(ctx, m)
			if updateErr := updateProxyStatus(ctx, m); updateErr != nil {
				err = updateErr
			}
		}
	}
	if result == nil {
		result = &ctrl.Result{}
	}

	m.Log.
		With("error", err).
		With("result", result).
		Info("reconciliation done")

	return *result, err
}

type StateMachineReconciler interface {
	Reconcile(ctx context.Context) (ctrl.Result, error)
}

func New(client client.Client, instance *v1alpha1.ImagePullReverseProxy, startState StateFn /*recorder record.EventRecorder,*/, scheme *apimachineryruntime.Scheme, log *zap.SugaredLogger, cache cache.BoolCache) StateMachineReconciler {
	sm := StateMachine{
		nextFn: startState,
		State: SystemState{
			ReverseProxy: *instance,
		},
		Log:    log,
		Client: client,
		Scheme: scheme,
		Cache:  cache,
	}
	sm.State.saveStatusSnapshot()
	return &sm
}

func updateProxyStatus(ctx context.Context, m *StateMachine) error {
	s := &m.State
	if !reflect.DeepEqual(s.ReverseProxy.Status, s.statusSnapshot) {
		m.Log.Debug(fmt.Sprintf("updating serverless status to '%+v'", s.ReverseProxy.Status))
		err := m.Client.Status().Update(ctx, &s.ReverseProxy)
		//emitEvent(r, s)
		s.saveStatusSnapshot()
		return err
	}
	return nil
}
