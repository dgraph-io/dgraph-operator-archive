/*
 * Copyright 2019-2020 Dgraph Labs, Inc. and Contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DgraphCluster is a Kubernetes custom resource which represents a
// dgraph cluster.
type DgraphCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the Dgraph cluster to create in the k8s cluster.
	Spec DgraphClusterSpec `json:"spec"`

	// Most recently observed status of the dgraph cluster
	Status DgraphClusterStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DgraphClusterList is the list of DgraphCluster in the k8s cluster.
type DgraphClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	// Items is the list of DgraphCluster
	Items []DgraphCluster `json:"items"`
}

// DgraphClusterSpec is the underlying specification of the DgraphCluster
// CRD.
type DgraphClusterSpec struct {
	// ClusterID is the ID of the dgraph cluster deployed.
	ClusterID string `json:"cluster_id"`
}

// DgraphClusterStatus represents the status of a DgraphCluster.
type DgraphClusterStatus struct {
	// Members is the Cilium policy status for each dgraph cluster member.
	Members map[string]DgraphClusterMemberStatus `json:"members,omitempty"`
}

// DgraphClusterMemberStatus represents the status of individual member in a
// dgraph cluster.
type DgraphClusterMemberStatus struct {
	// Healthy represents if the member is healthy or not.
	Healthy bool `json:"healthy"`
}
