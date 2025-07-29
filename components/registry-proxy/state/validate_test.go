package state

import (
	"context"
	"testing"

	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/api/v1alpha1"
	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/fsm"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_sFnValidateReverseProxyURL(t *testing.T) {
	t.Run("when function is valid should go to the next state", func(t *testing.T) {
		m := fsm.StateMachine{
			State: fsm.SystemState{
				Connection: v1alpha1.Connection{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "connection",
						Namespace: "maslo",
					},
					Spec: v1alpha1.ConnectionSpec{
						Proxy: v1alpha1.ConnectionSpecProxy{
							URL: "http://test-proxy-url",
						},
					},
				},
			},
		}

		next, result, err := sFnValidateReverseProxyURL(context.Background(), &m)

		// Assert
		// no errors
		require.Nil(t, err)
		// without stopping processing
		require.Nil(t, result)
		// with expected next state
		require.NotNil(t, next)
		requireEqualFunc(t, sFnConnectivityProxyURL, next)
		// function conditions remain unchanged
		require.Empty(t, m.State.Connection.Status.Conditions)
	})
	t.Run("when function is invalid should stop processing", func(t *testing.T) {
		m := fsm.StateMachine{
			State: fsm.SystemState{
				Connection: v1alpha1.Connection{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "connection",
						Namespace: "maslo",
					},
					Spec: v1alpha1.ConnectionSpec{
						Proxy: v1alpha1.ConnectionSpecProxy{
							URL: ":thisURLisbroken",
						},
					},
				},
			},
		}

		next, result, err := sFnValidateReverseProxyURL(context.Background(), &m)

		// Assert
		// no errors
		require.Nil(t, err)
		// without stopping processing
		require.Nil(t, result)
		// with expected next state
		require.Nil(t, next)
		// function has proper condition
		requireContainsCondition(t, m.State.Connection.Status,
			v1alpha1.ConditionConnectionReady,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonInvalidProxyURL,
			"Invalid Connectivity Proxy URL: parse \":thisURLisbroken\": missing protocol scheme")

	})
}
