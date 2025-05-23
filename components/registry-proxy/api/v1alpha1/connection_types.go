package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConnectionSpec defines the desired state of Connection.
type ConnectionSpec struct {
	// URL of the Connectivity Proxy, with protocol
	ProxyURL string `json:"proxyURL,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	TargetHost string                       `json:"targetHost"`
	Resources  *corev1.ResourceRequirements `json:"resources,omitempty"`
	// Sets desired log level to be used. The default value is "info"
	LogLevel string `json:"logLevel,omitempty"`
}

// ConnectionStatus defines the observed state of ConnectionStatus.
type ConnectionStatus struct {
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
	// connectivity proxy prerequisite
	ConditionConfigured ConditionType = "Configured"
)

type ConditionReason string

const (
	ConditionReasonDeploymentCreated            ConditionReason = "DeploymentCreated"
	ConditionReasonDeploymentUpdated            ConditionReason = "DeploymentUpdated"
	ConditionReasonDeploymentFailed             ConditionReason = "DeploymentFailed"
	ConditionReasonInvalidProxyURL              ConditionReason = "InvalidProxyURL"
	ConditionReasonProbeError                   ConditionReason = "ProbeError"
	ConditionReasonProbeSuccess                 ConditionReason = "ProbeSuccess"
	ConditionReasonProbeFailure                 ConditionReason = "ProbeFailure"
	ConditionReasonConnectivityProxyCrdUnknownn ConditionReason = "ConnectivityProxyCrdUnknown"
	ConditionReasonConnectivityProxyCrdFound    ConditionReason = "ConnectivityProxyCrdFound"
)

const (
	LabelApp        = "app"
	LabelModuleName = "kyma-project.io/module"
	LabelManagedBy  = "registry-proxy.kyma-project.io/managed-by"
	LabelResource   = "registry-proxy.kyma-project.io/resource"
	LabelName       = "app.kubernetes.io/name"
	LabelPartOf     = "app.kubernetes.io/part-of"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Running",type="string",JSONPath=".status.conditions[?(@.type=='Running')].status"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="NodePort",type="string",JSONPath=".status.nodePort"
// Connection is the Schema for the registryproxies API.
type Connection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConnectionSpec   `json:"spec,omitempty"`
	Status ConnectionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ConnectionList contains a list of Connection.
type ConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Connection `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Connection{}, &ConnectionList{})
}

func (connection *Connection) UpdateCondition(c ConditionType, s metav1.ConditionStatus, r ConditionReason, msg string) {
	condition := metav1.Condition{
		Type:               string(c),
		Status:             s,
		LastTransitionTime: metav1.Now(),
		Reason:             string(r),
		Message:            msg,
	}
	meta.SetStatusCondition(&connection.Status.Conditions, condition)
}
