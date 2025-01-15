package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ImagePullReverseProxySpec defines the desired state of ImagePullReverseProxy.
type ImagePullReverseProxySpec struct {
	// URL of the Connectivity Proxy, with protocol
	ProxyURL string `json:"proxyURL,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	TargetHost string `json:"targetHost"`
}

// ImagePullReverseProxyStatus defines the observed state of ImagePullReverseProxy.
type ImagePullReverseProxyStatus struct {
	// service nodeport number, then use localhost:<nodeport> to pull images
	NodePort int64 `json:"nodePort,omitempty"`

	// URL of the Connectivity Proxy
	ProxyURL string `json:"proxyURL,omitempty"`

	// Conditions associated with CustomStatus.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

type ConditionType string

const (
	ConditionDeployment      ConditionType = "Deployment"
	ConditionTargetReachable ConditionType = "TargetReachable"
)

type ConditionReason string

const (
	ConditionReasonDeploymentCreated ConditionReason = "DeploymentCreated"
	ConditionReasonDeploymentUpdated ConditionReason = "DeploymentUpdated"
	ConditionReasonDeploymentFailed  ConditionReason = "DeploymentFailed"
	ConditionReasonDeploymentWaiting ConditionReason = "DeploymentWaiting"
	ConditionReasonDeploymentReady   ConditionReason = "DeploymentReady"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Deployment",type="string",JSONPath=".status.conditions[?(@.type=='Deployment')].status"
// +kubebuilder:printcolumn:name="Target",type="string",JSONPath=".status.conditions[?(@.type=='TargetReachable')].status"
// ImagePullReverseProxy is the Schema for the imagepullreverseproxies API.
type ImagePullReverseProxy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ImagePullReverseProxySpec   `json:"spec,omitempty"`
	Status ImagePullReverseProxyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ImagePullReverseProxyList contains a list of ImagePullReverseProxy.
type ImagePullReverseProxyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ImagePullReverseProxy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ImagePullReverseProxy{}, &ImagePullReverseProxyList{})
}
