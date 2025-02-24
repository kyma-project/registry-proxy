package resources

import (
	"github.tools.sap/kyma/image-pull-reverse-proxy/components/controller/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type service struct {
	reverseProxy *v1alpha1.ImagePullReverseProxy
}

func NewService(rp *v1alpha1.ImagePullReverseProxy) *corev1.Service {
	s := &service{
		reverseProxy: rp,
	}
	return s.construct()
}

func (s *service) construct() *corev1.Service {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.reverseProxy.Name,
			Namespace: s.reverseProxy.Namespace,
			Labels:    labels(s.reverseProxy, "service"),
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{
				{
					Port:       80,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt(reverseProxyPort),
				},
			},
			Selector: map[string]string{
				v1alpha1.LabelApp: s.reverseProxy.Name,
			},
		},
	}

	return service
}
