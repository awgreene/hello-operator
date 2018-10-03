package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type HelloList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Hello `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Hello struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              HelloSpec   `json:"spec"`
	Status            HelloStatus `json:"status,omitempty"`
}

type HelloSpec struct {
	// Size is the size of the hello deployment
	Size  int32  `json:"size"`
	World string `json:"world"`
}
type HelloStatus struct {
	// Nodes are the names of the hello pods
	Nodes []string `json:"nodes"`
}
