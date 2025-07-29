package resources

import (
	"testing"

	"github.tools.sap/kyma/registry-proxy/components/common/container"
	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/api/v1alpha1"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewDeployment(t *testing.T) {
	t.Run("create deployment", func(t *testing.T) {
		rp := minimalConnection()

		d := NewDeployment(rp, rp.Spec.Proxy.URL, 0)

		require.NotNil(t, d)
		require.IsType(t, &appsv1.Deployment{}, d)
		require.Equal(t, "test-c-name", d.GetName())
		require.Equal(t, "test-c-namespace", d.GetNamespace())

		regContainer := container.Get(d.Spec.Template.Spec.Containers, RegistryContainerName)
		require.Contains(t, regContainer.Env, corev1.EnvVar{Name: "PROXY_URL", Value: "http://test-proxy-url"})
		require.Contains(t, regContainer.Env, corev1.EnvVar{Name: "TARGET_HOST", Value: "dummy"})

		require.Equal(t, defaultResources(), regContainer.Resources)
	})

	t.Run("create deployment with Resources", func(t *testing.T) {
		rp := minimalConnection()

		resources := minimalResources()
		rp.Spec.Resources = &resources

		d := NewDeployment(rp, rp.Spec.Proxy.URL, 0)

		require.NotNil(t, d)
		require.IsType(t, &appsv1.Deployment{}, d)
		require.Equal(t, "test-c-name", d.GetName())
		require.Equal(t, "test-c-namespace", d.GetNamespace())

		regContainer := container.Get(d.Spec.Template.Spec.Containers, RegistryContainerName)
		require.Equal(t, resources, regContainer.Resources)

		// TODO: contains in case of changed order
		require.Contains(t, regContainer.Env, corev1.EnvVar{Name: "PROXY_URL", Value: "http://test-proxy-url"})
		require.Contains(t, regContainer.Env, corev1.EnvVar{Name: "TARGET_HOST", Value: "dummy"})
	})

	t.Run("create deployment with authorizationHost", func(t *testing.T) {
		rp := minimalConnection()
		rp.Spec.Target.Authorization.Host = "example.com"

		d := NewDeployment(rp, rp.Spec.Proxy.URL, 123)

		require.NotNil(t, d)
		require.IsType(t, &appsv1.Deployment{}, d)
		require.Equal(t, "test-c-name", d.GetName())
		require.Equal(t, "test-c-namespace", d.GetNamespace())

		regContainer := container.Get(d.Spec.Template.Spec.Containers, RegistryContainerName)
		require.Contains(t, regContainer.Env, corev1.EnvVar{Name: "PROXY_URL", Value: "http://test-proxy-url"})
		require.Contains(t, regContainer.Env, corev1.EnvVar{Name: "TARGET_HOST", Value: "dummy"})
		require.Contains(t, regContainer.Env, corev1.EnvVar{Name: "AUTHORIZATION_NODE_PORT", Value: "123"})

		authContainer := container.Get(d.Spec.Template.Spec.Containers, AuthorizationContainerName)
		require.NotNil(t, authContainer)
		require.Contains(t, authContainer.Env, corev1.EnvVar{Name: "PROXY_URL", Value: "http://test-proxy-url"})
		require.Contains(t, authContainer.Env, corev1.EnvVar{Name: "TARGET_HOST", Value: "example.com"})

		require.Equal(t, defaultResources(), regContainer.Resources)
		require.Equal(t, defaultResources(), authContainer.Resources)
	})
}

func minimalConnection() *v1alpha1.Connection {
	return &v1alpha1.Connection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-c-name",
			Namespace: "test-c-namespace",
		},
		Spec: v1alpha1.ConnectionSpec{
			Proxy: v1alpha1.ConnectionSpecProxy{
				URL: "http://test-proxy-url",
			},
			Target: v1alpha1.ConnectionSpecTarget{
				Host: "dummy",
			},
		},
	}
}

func minimalRegistryProxyWithPort(desiredNodePort int32) *v1alpha1.Connection {
	return &v1alpha1.Connection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-rp-name",
			Namespace: "test-rp-namespace",
		},
		Spec: v1alpha1.ConnectionSpec{
			Proxy: v1alpha1.ConnectionSpecProxy{
				URL: "http://test-proxy-url",
			},
			Target: v1alpha1.ConnectionSpecTarget{
				Host: "dummy",
			},
			NodePort: desiredNodePort,
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
