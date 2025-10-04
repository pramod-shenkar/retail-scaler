/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OrgSpec defines the desired state of Org
type OrgSpec struct {
	OrgName         string        `json:"orgName"`
	Queue           Queue         `json:"queue"`
	Consumers       Consumers     `json:"consumers"`
	LagCheckerDelay time.Duration `json:"lagCheckerDelay"`
}

type Queue struct {
	Name      string     `json:"name"`
	Namespace string     `json:"namespace"`
	ClusterId string     `json:"clusterId"`
	Replicas  *int32     `json:"replicas"`
	Ports     QueuePorts `json:"ports"`
}

type QueuePorts struct {
	Plaintext  int32 `json:"plaintext"`
	Controller int32 `json:"controller"`
	External   int32 `json:"external"`
}

type Consumers struct {
	Namespace string  `json:"namespace"`
	GroupId   string  `json:"groupId"`
	Topics    []Topic `json:"topics"`
}

type Topic struct {
	Name           string `json:"name"`
	PartitionCount int    `json:"partitionCount"`
	MaxLag         int    `json:"maxLag"`
}

// OrgStatus defines the observed state of Org.
type OrgStatus struct {
	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	RunningConsumerCount int `json:"RunningConsumerCount,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Org is the Schema for the orgs API
type Org struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of Org
	// +required
	Spec OrgSpec `json:"spec"`

	// status defines the observed state of Org
	// +optional
	Status OrgStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// OrgList contains a list of Org
type OrgList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Org `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Org{}, &OrgList{})
}
