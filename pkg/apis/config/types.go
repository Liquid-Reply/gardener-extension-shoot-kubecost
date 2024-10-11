package config

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Configuration contains information about the rsyslog relp extension configuration.
type Configuration struct {
	metav1.TypeMeta

	ApiToken string
}
