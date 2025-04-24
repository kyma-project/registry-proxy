package resources

import (
	"testing"

	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/api/v1alpha1"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewDeployment(t *testing.T) {
	t.Run("create deployment", func(t *testing.T) {
		rp := minimalRegistryProxy()

		d := NewDeployment(rp, rp.Spec.ProxyURL)

		require.NotNil(t, d)
		require.IsType(t, &appsv1.Deployment{}, d)
		require.Equal(t, "test-rp-name", d.GetName())
		require.Equal(t, "test-rp-namespace", d.GetNamespace())
		require.Contains(t, d.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: "PROXY_URL", Value: "http://test-proxy-url"})
		require.Contains(t, d.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: "TARGET_HOST", Value: "dummy"})

		require.Equal(t, defaultResources(), d.Spec.Template.Spec.Containers[0].Resources)
	})

	t.Run("create deployment with Resources", func(t *testing.T) {
		rp := minimalRegistryProxy()

		resources := minimalResources()
		rp.Spec.Resources = &resources

		d := NewDeployment(rp, rp.Spec.ProxyURL)

		require.NotNil(t, d)
		require.IsType(t, &appsv1.Deployment{}, d)
		require.Equal(t, "test-rp-name", d.GetName())
		require.Equal(t, "test-rp-namespace", d.GetNamespace())

		require.Equal(t, resources, d.Spec.Template.Spec.Containers[0].Resources)

		require.Equal(t, "PROXY_URL", d.Spec.Template.Spec.Containers[0].Env[0].Name)
		require.Equal(t, "http://test-proxy-url", d.Spec.Template.Spec.Containers[0].Env[0].Value)
		require.Equal(t, "TARGET_HOST", d.Spec.Template.Spec.Containers[0].Env[1].Name)
		require.Equal(t, "dummy", d.Spec.Template.Spec.Containers[0].Env[1].Value)
	})
}

func minimalRegistryProxy() *v1alpha1.RegistryProxy {
	return &v1alpha1.RegistryProxy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-rp-name",
			Namespace: "test-rp-namespace",
		},
		Spec: v1alpha1.RegistryProxySpec{
			ProxyURL:   "http://test-proxy-url",
			TargetHost: "dummy",
		},
	}
}

func minimalResources() corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("128Mi"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("128Mi"),
		},
	}
}
