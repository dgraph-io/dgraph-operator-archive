/*
Copyright 2023.

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

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DgraphClusterSpec defines the desired state of DgraphCluster
type DgraphClusterSpec struct {
	// Size is the expected size of the Dgraph cluster.
	// The valid range of the size is from 1 to 7.
	Size int `json:"size"`

	// Repository is the name of the repository that hosts
	// Dgraph container images.
	//
	// By default, it is `docker.io/dgraph/dgraph`.
	Repository string `json:"repository,omitempty"`

	// Version is the expected version of the Dgraph cluster.
	//
	// Only released versions are supported: https://github.com/dgraph-io/dgraph/releases
	//
	// If version is not set, the default is "v23.1.0".
	Version string `json:"version,omitempty"`

	// AlphaPod defines the policy to create pod for the alpha pod.
	//
	// Updating Pod does not take effect on any existing alpha pods.
	AlphaPod *PodPolicy `json:"alphaPod,omitempty"`
	
	// ZeroPod defines the policy to create pod for the zero pod.
	//
	// Updating Pod does not take effect on any existing zero pods.
	ZeroPod *PodPolicy `json:"zeroPod,omitempty"`
}

// PodPolicy defines the policy to create pod for the Dgraph container.
type PodPolicy struct {
	// Shortest path for configuring the pod.
	// TO-DO: Abstract this out better.
	v1.PodSpec `json:",inline"`
	Labels map[string]string `json:"labels,omitempty"`
	PersistentVolumeClaimSpec *v1.PersistentVolumeClaimSpec `json:"persistentVolumeClaimSpec,omitempty"`
}

type ClusterPhase string

const (
	ClusterPhaseNone     ClusterPhase = ""
	ClusterPhaseCreating              = "Creating"
	ClusterPhaseRunning               = "Running"
	ClusterPhaseFailed                = "Failed"
)

type MembersStatus struct {
	// Ready are the dgraph nodes that are ready to serve requests
	// The member names are the same as the dgraph pod names
	Ready []string `json:"ready,omitempty"`
	// Unready are the dgraph nodes not ready to serve requests
	Unready []string `json:"unready,omitempty"`
}

// DgraphClusterStatus defines the observed state of DgraphCluster
type DgraphClusterStatus struct {
	// Phase is the cluster running phase
	Phase  ClusterPhase `json:"phase"`

	// Size is the current size of the cluster
	Size int `json:"size"`

	Alphas MembersStatus `json:"alphas"`
	Zeros MembersStatus `json:"zeros"`

	// CurrentVersion is the current cluster version
	CurrentVersion string `json:"currentVersion"`
	// TargetVersion is the version the cluster upgrading to.
	// If the cluster is not upgrading, TargetVersion is empty.
	TargetVersion string `json:"targetVersion"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DgraphCluster is the Schema for the dgraphclusters API
type DgraphCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DgraphClusterSpec   `json:"spec,omitempty"`
	Status DgraphClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DgraphClusterList contains a list of DgraphCluster
type DgraphClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DgraphCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DgraphCluster{}, &DgraphClusterList{})
}
