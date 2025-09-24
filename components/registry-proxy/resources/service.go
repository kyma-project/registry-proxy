package resources

import (
	"github.com/kyma-project/registry-proxy/components/registry-proxy/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type service struct {
	connection *v1alpha1.Connection
}

func NewService(connection *v1alpha1.Connection) *corev1.Service {
	s := &service{
		connection: connection,
	}
	return s.construct()
}

func (s *service) construct() *corev1.Service {
	port := corev1.ServicePort{
		Port:       80,
		Protocol:   corev1.ProtocolTCP,
		TargetPort: intstr.FromInt(registryProxyPort),
		Name:       RegistryContainerName,
	}
	if s.connection.Spec.NodePort != 0 {
		port.NodePort = s.connection.Spec.NodePort
	}
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.connection.Name,
			Namespace: s.connection.Namespace,
			Labels:    labels(s.connection, "service"),
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{
				port,
			},
			Selector: map[string]string{
				v1alpha1.LabelApp: s.connection.Name,
			},
		},
	}

	if s.connection.Spec.Target.Authorization.Host != "" {
		// TODO: tests
		authorizationPort := corev1.ServicePort{
			Port:       82,
			Protocol:   corev1.ProtocolTCP,
			TargetPort: intstr.FromInt(registryProxyAuthorizationPort),
			Name:       AuthorizationContainerName,
		}
		service.Spec.Ports = append(service.Spec.Ports, authorizationPort)
	}

	return service
}
