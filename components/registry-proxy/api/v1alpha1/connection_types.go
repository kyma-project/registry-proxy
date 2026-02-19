package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConnectionSpec defines the desired state of Connection.
type ConnectionSpec struct {
	// Details of the used proxy
	Proxy ConnectionSpecProxy `json:"proxy,omitempty"`

	// +kubebuilder:validation:Required
	Target    ConnectionSpecTarget         `json:"target"`
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// LogLevel sets the desired log level to be used.
	// Valid values are: "debug", "info", "warn", "error", "fatal".
	// The default value is "info".
	// +kubebuilder:validation:Enum=debug;info;warn;error;fatal
	// +kubebuilder:default=info
	LogLevel string `json:"logLevel,omitempty"`

	// NodePort is the port on which the service is exposed on each node.
	// If not specified, a random port will be assigned.
	NodePort int32 `json:"nodePort,omitempty,omitzero"`
}

type ConnectionSpecProxy struct {
	// URL of the Connectivity Proxy, with protocol
	URL string `json:"url,omitempty"`

	// Location ID of the connection
	// used to set the SAP-Connectivity-SCC-Location_ID header on every forwarded request
	LocationID string `json:"locationID,omitempty"`
}

type ConnectionSpecTarget struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Host string `json:"host"`

	// TODO: replace with AtMostOneOf in the future when it'll be available in kubebuilder: https://github.com/kubernetes-sigs/controller-tools/issues/461

	// Authorization defines the authorization method for the connection
	// +kubebuilder:validation:XValidation:message="Use host or headerSecret",rule="(!has(self.host) && !has(self.headerSecret)) || (has(self.host) && !has(self.headerSecret)) || (!has(self.host) && has(self.headerSecret))"
	Authorization ConnectionSpecTargetAuthorization `json:"authorization,omitempty"`
}

type ConnectionSpecTargetAuthorization struct {
	// Host is the name of the host that is used for registry authorization
	Host string `json:"host,omitempty"`

	// Name of the secret containing authorization header to be used for the connection
	HeaderSecret string `json:"headerSecret,omitempty"`
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
	ConditionConnectionDeployed ConditionType = "ConnectionDeployed"
	// pod readyz
	ConditionConnectionReady ConditionType = "ConnectionReady"
)

type ConditionReason string

const (
	ConditionReasonDeploymentCreated ConditionReason = "DeploymentCreated"
	ConditionReasonDeploymentUpdated ConditionReason = "DeploymentUpdated"
	ConditionReasonDeploymentFailed  ConditionReason = "DeploymentFailed"
	ConditionReasonInvalidProxyURL   ConditionReason = "InvalidProxyURL"
	ConditionReasonResourcesDeployed ConditionReason = "ConnectionResourcesDeployed"
	ConditionReasonResourcesNotReady ConditionReason = "ConnectionResourcesNotReady"
	ConditionReasonEstablished       ConditionReason = "ConnectionEstablished"
	ConditionReasonNotEstablished    ConditionReason = "ConnectionNotEstablished"
	ConditionReasonError             ConditionReason = "ConnectionError"
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
// +kubebuilder:printcolumn:name="Running",type="string",JSONPath=".status.conditions[?(@.type=='ConnectionDeployed')].status"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type=='ConnectionReady')].status"
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
