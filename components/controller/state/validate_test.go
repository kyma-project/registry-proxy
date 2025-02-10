package state

import (
	"context"
	"github.tools.sap/kyma/image-pull-reverse-proxy/components/controller/api/v1alpha1"
	"github.tools.sap/kyma/image-pull-reverse-proxy/components/controller/fsm"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_sFnValidateReverseProxyURL(t *testing.T) {
	t.Run("when function is valid should go to the next state", func(t *testing.T) {
		m := fsm.StateMachine{
			State: fsm.SystemState{
				ReverseProxy: v1alpha1.ImagePullReverseProxy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "rp",
						Namespace: "maslo",
					},
					Spec: v1alpha1.ImagePullReverseProxySpec{
						ProxyURL: "http://test-proxy-url",
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
		requireEqualFunc(t, sFnHandleDeployment, next)
		// function conditions remain unchanged
		require.Empty(t, m.State.ReverseProxy.Status.Conditions)
	})
	t.Run("when function is invalid should stop processing", func(t *testing.T) {
		m := fsm.StateMachine{
			State: fsm.SystemState{
				ReverseProxy: v1alpha1.ImagePullReverseProxy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "rp",
						Namespace: "maslo",
					},
					Spec: v1alpha1.ImagePullReverseProxySpec{
						ProxyURL: ":thisURLisbroken",
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
		requireContainsCondition(t, m.State.ReverseProxy.Status,
			v1alpha1.ConditionReady,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonInvalidProxyURL,
			"Invalid Connectivity Proxy URL: parse \":thisURLisbroken\": missing protocol scheme")

	})
}
