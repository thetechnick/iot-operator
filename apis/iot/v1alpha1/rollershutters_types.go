package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Position",type="number",JSONPath=".status.position"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type RollerShutter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RollerShutterSpec   `json:"spec,omitempty"`
	Status RollerShutterStatus `json:"status,omitempty"`
}

type RollerShutterSpec struct {
	// Endpoint device type.
	DeviceType string                `json:"deviceType"`
	Endpoint   RollerShutterEndpoint `json:"endpoint"`
	// Desired position for the shutter.
	Position int `json:"position"`
}

type RollerShutterEndpoint struct {
	// URL to contact the device under.
	URL string `json:"url"`
}

type RollerShutterStatus struct {
	// The most recent generation observed by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions is a list of status conditions ths object is in.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	Phase      RollerShutterPhase `json:"phase,omitempty"`
	// Recorded position in percentage open.
	// 100 = completely open, 0 = completely closed.
	Position int `json:"position"`
	// Power consumption in Watts.
	Power int `json:"power"`
}

const (
	// Condition indicating whether the device can be contacted
	RollerShutterReachable = "Reachable"
	// Condition indicating whether the shutter is at the commanded position
	RollerShutterAtPosition = "AtPosition"
)

type RollerShutterPhase string

const (
	RollerShutterPhaseClosing RollerShutterPhase = "Closing"
	RollerShutterPhaseIdle    RollerShutterPhase = "Idle"
	RollerShutterPhaseOpening RollerShutterPhase = "Opening"
)

// RollerShutterList contains a list of RollerShutters
// +kubebuilder:object:root=true
type RollerShutterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RollerShutter `json:"items"`
}

func init() {
	register(&RollerShutter{}, &RollerShutterList{})
}
