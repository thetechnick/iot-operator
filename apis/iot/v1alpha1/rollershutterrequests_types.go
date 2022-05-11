package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Position",type="number",JSONPath=".status.position"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type RollerShutterRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RollerShutterRequestSpec   `json:"spec,omitempty"`
	Status RollerShutterRequestStatus `json:"status,omitempty"`
}

type RollerShutterRequestSpec struct {
	// Desired position for the shutter.
	Position      int                         `json:"position"`
	RollerShutter corev1.LocalObjectReference `json:"rollerShutter"`
}

type RollerShutterRequestStatus struct {
	// The most recent generation observed by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions is a list of status conditions ths object is in.
	Conditions []metav1.Condition        `json:"conditions,omitempty"`
	Phase      RollerShutterRequestPhase `json:"phase,omitempty"`
}

const (
	// Condition indicating whether the request was completed
	RollerShutterRequestCompleted = "Completed"
)

type RollerShutterRequestPhase string

const (
	RollerShutterRequestPhasePending   RollerShutterRequestPhase = "Pending"
	RollerShutterRequestPhaseMoving    RollerShutterRequestPhase = "Moving"
	RollerShutterRequestPhaseCompleted RollerShutterRequestPhase = "Completed"
)

// RollerShutterRequestList contains a list of RollerShutterRequests
// +kubebuilder:object:root=true
type RollerShutterRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RollerShutterRequest `json:"items"`
}

func init() {
	register(&RollerShutterRequest{}, &RollerShutterRequestList{})
}
