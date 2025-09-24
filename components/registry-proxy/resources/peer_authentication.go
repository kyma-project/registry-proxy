package resources

import (
	"github.com/kyma-project/registry-proxy/components/registry-proxy/api/v1alpha1"
	apisecurityv1 "istio.io/api/security/v1"
	apitypev1beta1 "istio.io/api/type/v1beta1"
	securityclientv1 "istio.io/client-go/pkg/apis/security/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type peerAuthentication struct {
	connection *v1alpha1.Connection
}

func NewPeerAuthentication(connection *v1alpha1.Connection) *securityclientv1.PeerAuthentication {
	p := &peerAuthentication{
		connection: connection,
	}
	return p.construct()
}

func (p *peerAuthentication) construct() *securityclientv1.PeerAuthentication {
	pa := &securityclientv1.PeerAuthentication{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.connection.Name,
			Namespace: p.connection.Namespace,
			Labels:    labels(p.connection, "peer-authentication"),
		},
		Spec: apisecurityv1.PeerAuthentication{
			Selector: &apitypev1beta1.WorkloadSelector{
				MatchLabels: map[string]string{
					v1alpha1.LabelApp: p.connection.Name,
				},
			},
			Mtls: &apisecurityv1.PeerAuthentication_MutualTLS{
				Mode: apisecurityv1.PeerAuthentication_MutualTLS_PERMISSIVE,
			},
		},
	}
	return pa
}
