package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RegistryProxySpec defines the desired state of RegistryProxy.
type RegistryProxySpec struct {
	// TODO: think if we want any global fields here, like log level?
}

type State string
type Served string
type ConditionType string
type ConditionReason string

// TODO: add condition types, reasons, labels

const (
	StateReady      State = "Ready"
	StateProcessing State = "Processing"
	StateWarning    State = "Warning"
	StateError      State = "Error"
	StateDeleting   State = "Deleting"

	ServedTrue  Served = "True"
	ServedFalse Served = "False"

	// installation and deletion details
	ConditionTypeInstalled = ConditionType("Installed")

	// prerequisites and soft dependencies
	ConditionTypeConfigured = ConditionType("Configured")

	// deletion
	ConditionTypeDeleted = ConditionType("Deleted")

	ConditionReasonConfiguration           = ConditionReason("Configuration")
	ConditionReasonConfigurationErr        = ConditionReason("ConfigurationErr")
	ConditionReasonConfigured              = ConditionReason("Configured")
	ConditionReasonInstallation            = ConditionReason("Installation")
	ConditionReasonInstallationErr         = ConditionReason("InstallationErr")
	ConditionReasonInstalled               = ConditionReason("Installed")
	ConditionReasonRegistryProxyDuplicated = ConditionReason("RegistryProxyDuplicated")
	ConditionReasonDeletion                = ConditionReason("Deletion")
	ConditionReasonDeletionErr             = ConditionReason("DeletionErr")
	ConditionReasonDeleted                 = ConditionReason("Deleted")

	Finalizer = "registry-proxy-operator.kyma-project.io/deletion-hook"
)

// RegistryProxyStatus defines the observed state of RegistryProxy.
type RegistryProxyStatus struct {

	// State signifies current state of RegistryProxy.
	// Value can be one of ("Ready", "Processing", "Error", "Deleting").
	// +kubebuilder:validation:Enum=Processing;Deleting;Ready;Error;Warning
	State State `json:"state,omitempty"`

	// Served signifies that current RegistryProxy is managed.
	// Value can be one of ("True", "False").
	// +kubebuilder:validation:Enum=True;False
	Served Served `json:"served"`

	// TODO: status, or maybe conditions is enough?
	// Conditions associated with CustomStatus.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// TODO add columns to print
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// RegistryProxy is the Schema for the RegistryProxies API.
type RegistryProxy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RegistryProxySpec   `json:"spec,omitempty"`
	Status RegistryProxyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RegistryProxyList contains a list of RegistryProxy.
type RegistryProxyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RegistryProxy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RegistryProxy{}, &RegistryProxyList{})
}

// TODO: check if we can use some kind of generic here, alongisde with Conenction CRD
func (rp *RegistryProxy) UpdateCondition(c ConditionType, s metav1.ConditionStatus, r ConditionReason, msg string) {
	condition := metav1.Condition{
		Type:               string(c),
		Status:             s,
		LastTransitionTime: metav1.Now(),
		Reason:             string(r),
		Message:            msg,
	}
	meta.SetStatusCondition(&rp.Status.Conditions, condition)
}

func (s *RegistryProxy) IsServedEmpty() bool {
	return s.Status.Served == ""
}
