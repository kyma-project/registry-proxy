package state

import (
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/api/v1alpha1"
	v1alpha2 "github.tools.sap/kyma/registry-proxy/components/registry-proxy/api/v1alpha1"
	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/fsm"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// requireEqualFunc compares two stateFns based on their names returned from the reflect package
// names returned from the package may be different for any go/dlv compiler version
// for go1.22 returned name is in format:
// github.com/kyma-project/keda-manager/pkg/reconciler.Test_sFnServedFilter.func4.sFnUpdateStatus.3
func requireEqualFunc(t *testing.T, expected, actual fsm.StateFn) {
	expectedFnName := getFnName(expected)
	actualFnName := getFnName(actual)

	if expectedFnName == actualFnName {
		// return if functions are simply same
		return
	}

	expectedElems := strings.Split(expectedFnName, "/")
	actualElems := strings.Split(actualFnName, "/")

	// check package paths (prefix)
	// e.g. 'github.com/kyma-project/keda-manager/pkg'
	require.Equal(t,
		strings.Join(expectedElems[0:len(expectedElems)-2], "/"),
		strings.Join(actualElems[0:len(actualElems)-2], "/"),
	)

	// check direct fn names (suffix)
	// e.g. 'reconciler.Test_sFnServedFilter.func4.sFnUpdateStatus.3'
	require.Equal(t,
		getDirectFnName(expectedElems[len(expectedElems)-1]),
		getDirectFnName(actualElems[len(actualElems)-1]),
	)
}

func getDirectFnName(nameSuffix string) string {
	elements := strings.Split(nameSuffix, ".")
	return elements[len(elements)-2]
}

func getFnName(fn fsm.StateFn) string {
	return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
}

func requireContainsCondition(t *testing.T, status v1alpha1.ImagePullReverseProxyStatus,
	conditionType v1alpha1.ConditionType, conditionStatus metav1.ConditionStatus,
	conditionReason v1alpha1.ConditionReason, conditionMessage string) {
	hasExpectedCondition := false
	for _, condition := range status.Conditions {
		if condition.Type == string(conditionType) {
			require.Equal(t, string(conditionReason), condition.Reason)
			require.Equal(t, conditionStatus, condition.Status)
			require.Equal(t, conditionMessage, condition.Message)
			hasExpectedCondition = true
		}
	}
	require.True(t, hasExpectedCondition)
}

func minimalScheme(t *testing.T) *k8sruntime.Scheme {
	scheme := k8sruntime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, v1alpha2.AddToScheme(scheme))
	require.NoError(t, appsv1.AddToScheme(scheme))
	return scheme
}
