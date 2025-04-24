package resources

import (
	"github.tools.sap/kyma/registry-proxy/components/registry-proxy/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type service struct {
	registryProxy *v1alpha1.RegistryProxy
}

func NewService(rp *v1alpha1.RegistryProxy) *corev1.Service {
	s := &service{
		registryProxy: rp,
	}
	return s.construct()
}

func (s *service) construct() *corev1.Service {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.registryProxy.Name,
			Namespace: s.registryProxy.Namespace,
			Labels:    labels(s.registryProxy, "service"),
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{
				{
					Port:       80,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt(registryProxyPort),
				},
			},
			Selector: map[string]string{
				v1alpha1.LabelApp: s.registryProxy.Name,
			},
		},
	}

	return service
}
