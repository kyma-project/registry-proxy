package resources

import (
	"fmt"
	"os"
	"strconv"

	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/api/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

const (
	defaultLimitCPU                = "100m"
	defaultLimitMemory             = "64Mi"
	defaultRequestCPU              = "5m"
	defaultRequestMemory           = "32Mi"
	registryProxyPort              = 8080
	probesPort                     = 8081
	registryProxyAuthorizationPort = 8082
	authorizationProbesPort        = 8083
	// TODO: move to some common resources package?
	RegistryContainerName      = "registry"
	AuthorizationContainerName = "authorization"
)

type deployment struct {
	connection            *v1alpha1.Connection
	proxyURL              string
	authorizationNodePort int32
}

func NewDeployment(connection *v1alpha1.Connection, proxyURL string, authorizationNodePort int32) *appsv1.Deployment {
	d := &deployment{
		connection:            connection,
		proxyURL:              proxyURL,
		authorizationNodePort: authorizationNodePort,
	}
	return d.construct()
}

func (d *deployment) construct() *appsv1.Deployment {
	podSelectorLabels := labels(d.connection, "deployment")
	podLabels := labels(d.connection, "deployment")
	podLabels["sidecar.istio.io/inject"] = "true"

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.connection.Name,
			Namespace: d.connection.Namespace,
			Labels:    podSelectorLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: podSelectorLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: podLabels,
				},
				Spec: corev1.PodSpec{
					Containers: d.containers(),
				},
			},
			Replicas: ptr.To[int32](1),
		},
	}
	if d.connection.Spec.Target.Authorization.HeaderSecret != "" {
		deployment.Spec.Template.Spec.Volumes = []corev1.Volume{
			{
				Name: "authorization",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: d.connection.Spec.Target.Authorization.HeaderSecret,
					},
				},
			},
		}
	}
	return deployment
}

func (d *deployment) containers() []corev1.Container {
	containers := make([]corev1.Container, 0)
	envs := d.envs()
	registryContainer := d.container(RegistryContainerName, registryProxyPort, probesPort, envs)
	if d.connection.Spec.Target.Authorization.HeaderSecret != "" {
		registryContainer.VolumeMounts = []corev1.VolumeMount{
			{
				Name:      "authorization",
				MountPath: "/secrets/authorization",
				ReadOnly:  true,
			},
		}
	}
	containers = append(containers, registryContainer)

	if d.authorizationNodePort != 0 {
		authorizationEnvs := d.authEnvs()
		authorizationContainer := d.container(AuthorizationContainerName, registryProxyAuthorizationPort, authorizationProbesPort, authorizationEnvs)
		containers = append(containers, authorizationContainer)
	}

	return containers
}

func (d *deployment) container(name string, port, probePort int32, envs []corev1.EnvVar) corev1.Container {
	container := corev1.Container{
		Name:  name,
		Image: os.Getenv("PROXY_IMAGE"),
		Command: []string{
			os.Getenv("PROXY_COMMAND"),
		},
		Args: []string{
			"--connection-bind-address", fmt.Sprintf(":%d", port),
			"--health-probe-bind-address", fmt.Sprintf(":%d", probePort),
		},
		ImagePullPolicy: corev1.PullIfNotPresent,
		Resources:       d.resourceConfiguration(),
		Env:             envs,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: port,
				Protocol:      "TCP",
			},
		},
		StartupProbe: &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/healthz",
					Port: intstr.FromInt32(probePort),
				},
			},
			InitialDelaySeconds: 0,
			PeriodSeconds:       5,
			SuccessThreshold:    1,
			FailureThreshold:    30, // FailureThreshold * PeriodSeconds = 150s in this case, this should be enough for any function pod to start up
		},
		ReadinessProbe: &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/readyz",
					Port: intstr.FromInt32(probePort),
				},
			},
			InitialDelaySeconds: 0, // startup probe exists, so delaying anything here doesn't make sense
			FailureThreshold:    1,
			PeriodSeconds:       10,
			TimeoutSeconds:      2,
		},
		LivenessProbe: &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/healthz",
					Port: intstr.FromInt32(probePort),
				},
			},
			FailureThreshold: 3,
			PeriodSeconds:    5,
			TimeoutSeconds:   4,
		},
		SecurityContext: &corev1.SecurityContext{
			RunAsGroup: d.podRunAsUserUID(), // set to 1000 because default value is root(0)
			RunAsUser:  d.podRunAsUserUID(),
			SeccompProfile: &corev1.SeccompProfile{
				Type: corev1.SeccompProfileTypeRuntimeDefault,
			},
			AllowPrivilegeEscalation: ptr.To(false),
			RunAsNonRoot:             ptr.To(true),
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{
					"All",
				},
			},
		},
	}

	return container
}

func (d *deployment) resourceConfiguration() corev1.ResourceRequirements {
	if d.connection.Spec.Resources != nil {
		return *d.connection.Spec.Resources
	}

	return defaultResources()
}

func defaultResources() corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(defaultLimitCPU),
			corev1.ResourceMemory: resource.MustParse(defaultLimitMemory),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(defaultRequestCPU),
			corev1.ResourceMemory: resource.MustParse(defaultRequestMemory),
		},
	}
}

func (d *deployment) envs() []corev1.EnvVar {
	envVariables := []corev1.EnvVar{
		{
			Name:  "PROXY_URL",
			Value: d.proxyURL,
		},
		{
			Name:  "TARGET_HOST",
			Value: d.connection.Spec.Target.Host,
		},
		// TOOD: maybe skip setting this if empty?
		{
			Name:  "LOCATION_ID",
			Value: d.connection.Spec.Proxy.LocationID,
		},
	}

	if d.connection.Spec.LogLevel != "" {
		envVariables = append(envVariables, corev1.EnvVar{
			Name:  "LOG_LEVEL",
			Value: d.connection.Spec.LogLevel,
		})
	}

	if d.authorizationNodePort != 0 {
		envVariables = append(envVariables, corev1.EnvVar{
			Name:  "AUTHORIZATION_NODE_PORT",
			Value: strconv.Itoa(int(d.authorizationNodePort)),
		})
	}

	return envVariables
}

func (d *deployment) authEnvs() []corev1.EnvVar {
	envVariables := []corev1.EnvVar{
		{
			Name:  "PROXY_URL",
			Value: d.proxyURL,
		},
		{
			Name:  "TARGET_HOST",
			Value: d.connection.Spec.Target.Authorization.Host,
		},
		{
			Name:  "LOCATION_ID",
			Value: d.connection.Spec.Proxy.LocationID,
		},
	}

	if d.connection.Spec.LogLevel != "" {
		envVariables = append(envVariables, corev1.EnvVar{
			Name:  "LOG_LEVEL",
			Value: d.connection.Spec.LogLevel,
		})
	}

	return envVariables
}

func (d *deployment) podRunAsUserUID() *int64 {
	return ptr.To[int64](1000) // runAsUser 1000 is the most popular and standard value for non-root user
}
