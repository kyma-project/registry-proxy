package connectivityproxy

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type ConnectivityProxy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec ConnectivityProxySpec `json:"spec"`
}

type ConnectivityProxySpec struct {
	Config ConnectivityProxyConfig `json:"config"`
}

type ConnectivityProxyConfig struct {
	Servers ConnectivityProxyServers `json:"servers"`
}

type ConnectivityProxyServers struct {
	Proxy ConnectivityProxyServerProxy `json:"proxy"`
}

type ConnectivityProxyServerProxy struct {
	Http ConnectivityProxyHttp `json:"http,omitempty"`
}

type ConnectivityProxyHttp struct {
	Port int `json:"port,omitempty"`
}

// +kubebuilder:object:root=true

// ConnectivityProxyList contains a list of ConnectivityProxy.
type ConnectivityProxyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConnectivityProxy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ConnectivityProxy{}, &ConnectivityProxyList{})
}
