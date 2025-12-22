package fsm

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/kyma-project/registry-proxy/components/common/cache"

	"github.com/kyma-project/manager-toolkit/installation/chart"
	"github.com/kyma-project/registry-proxy/components/operator/api/v1alpha1"
	"github.com/kyma-project/registry-proxy/components/operator/flags"

	"go.uber.org/zap"
	apimachineryruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	chartPath = "/module-chart"
)

var (
	secretCacheKey = types.NamespacedName{
		Name:      "registry-proxy-manifest-cache",
		Namespace: "registry-proxy",
	}
)

type StateFn func(context.Context, *StateMachine) (StateFn, *ctrl.Result, error)

type SystemState struct {
	RegistryProxy  v1alpha1.RegistryProxy
	statusSnapshot v1alpha1.RegistryProxyStatus
	ChartConfig    *chart.Config
	Cache          chart.ManifestCache
	FlagsBuilder   *flags.Builder
}

func (s *SystemState) saveStatusSnapshot() {
	result := s.RegistryProxy.Status.DeepCopy()
	if result == nil {
		result = &v1alpha1.RegistryProxyStatus{}
	}
	s.statusSnapshot = *result
}

// TODO: think if we can use generics here to have one state machine for both RegistryProxy and Connection
type StateMachine struct {
	nextFn                     StateFn
	State                      SystemState
	Log                        *zap.SugaredLogger
	Client                     client.Client
	Scheme                     *apimachineryruntime.Scheme
	ConnectivityProxyReadiness cache.BoolCache
	IstioReadiness             cache.BoolCache
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

func New(client client.Client, config *rest.Config, instance *v1alpha1.RegistryProxy, startState StateFn, scheme *apimachineryruntime.Scheme, log *zap.SugaredLogger, connectivityProxyReadiness cache.BoolCache, istioReadiness cache.BoolCache, chartCache chart.ManifestCache) StateMachineReconciler {
	sm := StateMachine{
		nextFn: startState,
		State: SystemState{
			RegistryProxy: *instance,
			ChartConfig:   chartConfig(context.Background(), client, config, log, chartCache, instance.Namespace),
			FlagsBuilder:  flags.NewBuilder(),
		},
		Log:                        log,
		Client:                     client,
		Scheme:                     scheme,
		ConnectivityProxyReadiness: connectivityProxyReadiness,
		IstioReadiness:             istioReadiness,
	}
	sm.State.saveStatusSnapshot()
	return &sm
}

func updateProxyStatus(ctx context.Context, m *StateMachine) error {
	s := &m.State
	if !reflect.DeepEqual(s.RegistryProxy.Status, s.statusSnapshot) {
		m.Log.Debug(fmt.Sprintf("updating image pull registry proxy status to '%+v'", s.RegistryProxy.Status))
		err := m.Client.Status().Update(ctx, &s.RegistryProxy)
		//emitEvent(r, s)
		s.saveStatusSnapshot()
		return err
	}
	return nil
}

func chartConfig(ctx context.Context, client client.Client, config *rest.Config, log *zap.SugaredLogger, cache chart.ManifestCache, namespace string) *chart.Config {
	return &chart.Config{
		Ctx:         ctx,
		Log:         log,
		Cache:       cache,
		CacheKey:    secretCacheKey,
		ManagerUID:  os.Getenv("REGISTRYPROXY_MANAGER_UID"),
		ManagerName: "registry-proxy-operator",
		Cluster: chart.Cluster{
			Client: client,
			Config: config,
		},
		Release: chart.Release{
			ChartPath: chartPath,
			Namespace: namespace,
			Name:      "registry-proxy",
		},
	}
}
