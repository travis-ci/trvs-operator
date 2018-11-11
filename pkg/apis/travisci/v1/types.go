package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type TrvsSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              TrvsSecretSpec `json:"spec"`
}

type TrvsSecretSpec struct {
	App         string `json:"app"`
	Environment string `json:"env"`
	Prefix      string `json:"prefix"`
	IsPro       bool   `json:"pro"`
	File        string `json:"file"`
	Key         string `json:"key"`
	RawKeys     bool   `json:"rawKeys"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type TrvsSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TrvsSecret `json:"items"`
}
