package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:openapi-gen=true

type TestSpec struct {

	// Version is the ELK version to use
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Version string `json:"version"`
}

// +kubebuilder:object:root=true
type XTest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              TestSpec `json:"spec,omitempty"`
}
