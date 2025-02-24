package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ImagePullReverseProxySpec defines the desired state of ImagePullReverseProxy.
type ImagePullReverseProxySpec struct {
	// URL of the Connectivity Proxy, with protocol
	ProxyURL string `json:"proxyURL,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	TargetHost string                       `json:"targetHost"`
	Resources  *corev1.ResourceRequirements `json:"resources,omitempty"`
	// Sets desired log level to be used. The default value is "info"
	LogLevel string `json:"logLevel,omitempty"`
}

// ImagePullReverseProxyStatus defines the observed state of ImagePullReverseProxy.
type ImagePullReverseProxyStatus struct {
	// service nodeport number, then use localhost:<nodeport> to pull images
	NodePort int32 `json:"nodePort,omitempty,omitzero"`

	// URL of the Connectivity Proxy
	ProxyURL string `json:"proxyURL,omitempty,omitzero"`

	// Conditions associated with CustomStatus.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

type ConditionType string

const (
	// pod healthz
	ConditionRunning ConditionType = "Running"
	// pod readyz
	ConditionReady ConditionType = "Ready"
)

type ConditionReason string

const (
	ConditionReasonDeploymentCreated ConditionReason = "DeploymentCreated"
	ConditionReasonDeploymentUpdated ConditionReason = "DeploymentUpdated"
	ConditionReasonDeploymentFailed  ConditionReason = "DeploymentFailed"
	ConditionReasonDeploymentWaiting ConditionReason = "DeploymentWaiting"
	ConditionReasonDeploymentReady   ConditionReason = "DeploymentReady"
	ConditionReasonInvalidProxyURL   ConditionReason = "InvalidProxyURL"
	ConditionReasonProbeError        ConditionReason = "ProbeError"
	ConditionReasonProbeSuccess      ConditionReason = "ProbeSuccess"
	ConditionReasonProbeFailure      ConditionReason = "ProbeFailure"
)

const (
	LabelApp        = "app"
	LabelModuleName = "kyma-project.io/module"
	LabelManagedBy  = "image-pull-reverse-proxy.kyma-project.io/managed-by"
	LabelResource   = "image-pull-reverse-proxy.kyma-project.io/resource"
	LabelName       = "app.kubernetes.io/name"
	LabelPartOf     = "app.kubernetes.io/part-of"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Running",type="string",JSONPath=".status.conditions[?(@.type=='Running')].status"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="NodePort",type="string",JSONPath=".status.nodePort"
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

func (rp *ImagePullReverseProxy) UpdateCondition(c ConditionType, s metav1.ConditionStatus, r ConditionReason, msg string) {
	condition := metav1.Condition{
		Type:               string(c),
		Status:             s,
		LastTransitionTime: metav1.Now(),
		Reason:             string(r),
		Message:            msg,
	}
	meta.SetStatusCondition(&rp.Status.Conditions, condition)
}
