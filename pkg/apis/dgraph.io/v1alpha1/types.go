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
	k8sutils "github.com/dgraph-io/dgraph-operator/pkg/k8s/utils"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterState represents the state of the cluster.
type ClusterState string

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +k8s:openapi-gen=true
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

// AsOwnerReference returns the OwnerReference corresponding to DgraphCluster
// which can be used as OwnerReference for other resources in the cluster.
func (dc *DgraphCluster) AsOwnerReference() metav1.OwnerReference {
	controller := true
	blockOwnerDeletion := true

	return metav1.OwnerReference{
		APIVersion:         SchemeGroupVersion.String(),
		Kind:               DgraphClusterKindDefinition,
		Name:               dc.GetName(),
		UID:                dc.GetUID(),
		Controller:         &controller,
		BlockOwnerDeletion: &blockOwnerDeletion,
	}
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +k8s:openapi-gen=true
// DgraphClusterList is the list of DgraphCluster in the k8s cluster.
type DgraphClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	// Items is the list of DgraphCluster
	Items []DgraphCluster `json:"items"`
}

// +k8s:openapi-gen=true
// DgraphClusterSpec is the underlying specification of the DgraphCluster
// CRD.
// There are three important components of a Dgraph Cluster
// 1. Alpha
// 2. Zero
// 3. Ratel(optional)
type DgraphClusterSpec struct {
	// ClusterID is the ID of the dgraph cluster deployed.
	ClusterID string `json:"clusterID"`

	// Cluster specification for dgraph alpha components.
	AlphaCluster AlphaClusterSpec `json:"alpha"`

	// Cluster specification for dgraph zero components.
	ZeroCluster ZeroClusterSpec `json:"zero"`

	// Specification for dgraph ratel component for providing UI.
	Ratel RatelSpec `json:"ratel,omitempty"`

	// Below variables are more or less same
	// as that of DgraphComponentSpec but are cluster level, they can be overridden
	// inside individual component configuration.

	// Base image of the component
	BaseImage string `json:"baseImage"`

	// Version of the component. Override the cluster-level version if non-empty
	Version string `json:"version"`

	// ServiceType is the type of kubernetes service to create for the Cluster components.
	ServiceType string `jsong:"serviceType,omitempty"`

	// ImagePullPolicy of the dgraph component.
	ImagePullPolicy *corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// Annotations of the component.
	// Cluster level annotation is not overridden by the component configuration
	// rather merged with the underlying specified annotations.
	Annotations map[string]string `json:"annotations,omitempty"`

	// Resource requirements of the components, this can be overridden at component level.
	corev1.ResourceRequirements `json:",inline"`
}

// AlphaServiceType returns the kubernetes service type to use for Alpha Cluster
func (dc *DgraphClusterSpec) AlphaServiceType() corev1.ServiceType {
	if dc.AlphaCluster.ServiceType != "" {
		return k8sutils.ResolveK8SServiceType(dc.AlphaCluster.ServiceType)
	}

	return k8sutils.ResolveK8SServiceType(dc.ServiceType)
}

// ZeroServiceType returns the kubernetes service type to use for Zero Cluster
func (dc *DgraphClusterSpec) ZeroServiceType() corev1.ServiceType {
	if dc.ZeroCluster.ServiceType != "" {
		return k8sutils.ResolveK8SServiceType(dc.ZeroCluster.ServiceType)
	}

	return k8sutils.ResolveK8SServiceType(dc.ServiceType)
}

// RatelServiceType returns the kubernetes service type to use for Ratel Cluster
func (dc *DgraphClusterSpec) RatelServiceType() corev1.ServiceType {
	if dc.Ratel.ServiceType != "" {
		return k8sutils.ResolveK8SServiceType(dc.Ratel.ServiceType)
	}

	return k8sutils.ResolveK8SServiceType(dc.ServiceType)
}

// GetClusterID returns the cluster ID for the provided Dgraph Cluster.
func (dcs *DgraphClusterSpec) GetClusterID() string {
	return dcs.ClusterID
}

// DgraphClusterStatus represents the status of a DgraphCluster.
type DgraphClusterStatus struct {
	// ClusterID is the ID of the dgraph cluster deployed.
	ClusterID string `json:"cluster_id"`

	State ClusterState `json:"state,omitempty"`

	// Status of individual dgraph components like alpha, zero and ratel.
	AlphaCluster AlphaClusterStatus `json:"alpha,omitempty"`
	ZeroCluster  ZeroClusterStatus  `json:"zero,omitempty"`
	Ratel        RatelStatus        `json:"ratel,omitempty"`
}

// +k8s:openapi-gen=true
// AlphaClusterSpec is the specification of the dgraph alpha cluster.
type AlphaClusterSpec struct {
	DgraphComponentSpec `json:",inline"`

	// Number of replicas to run in the cluster.
	Replicas int32 `json:"replicas"`

	// StorageClass to use as persistent volume for the componnet.
	StorageClassName string `json:"storageClassName,omitempty"`

	// Config is the configuration of the dgraph component.
	Config *AlphaConfig `json:"config,omitempty"`
}

// AlphaClusterStatus represents the cluster status of dgraph alpha components.
type AlphaClusterStatus struct {
	// StatefulSet is the status of stateful set associated with the specified
	// alpha cluster.
	StatefulSet *apps.StatefulSetStatus `json:"statefulSet,omitempty"`

	// Members is the map of members in the alpha cluster.
	Members map[string]DgraphComponent `json:"members,omitempty"`
}

// +k8s:openapi-gen=true
// ZeroClusterSpec is the specification of the dgraph alpha cluster.
type ZeroClusterSpec struct {
	DgraphComponentSpec `json:",inline"`

	// Number of replicas to run in the cluster.
	Replicas int32 `json:"replicas"`

	// StorageClass to use as persistent volume for the componnet.
	StorageClassName string `json:"storageClassName,omitempty"`

	// Config is the configuration of the dgraph zero.
	Config *ZeroConfig `json:"config,omitempty"`
}

// ZeroClusterStatus represents the cluster status of dgraph alpha components.
type ZeroClusterStatus struct {
	// StatefulSet is the status of stateful set associated with the specified
	// zero cluster.
	StatefulSet *apps.StatefulSetStatus `json:"statefulSet,omitempty"`

	// Members is the map of members in the zero cluster.
	Members map[string]DgraphComponent `json:"members,omitempty"`
}

// +k8s:openapi-gen=true
// RatelSpec holds the configuration of dgraph ratel components.
type RatelSpec struct {
	DgraphComponentSpec `json:",inline"`

	// Number of replicas of ratel to run in the cluster.
	Replicas int32 `json:"replicas"`
}

// RatelStatus holds the status of dgraph ratel component.
type RatelStatus struct {
}

// +k8s:openapi-gen=true
// DgraphComponentSpec is the common configuration values shared among different
// dgraph components.
type DgraphComponentSpec struct {
	// Resource requirements of the components.
	corev1.ResourceRequirements `json:",inline"`

	// Base image of the component
	BaseImage string `json:"baseImage,omitempty"`

	// ServiceType is type of service to create for the component.
	// One of NodePort, ClusterIP, LoadBalancer. Defaults to ClusterIP.
	ServiceType string `json:"serviceType,omitempty"`

	// Version of the component. Override the cluster-level version if non-empty
	Version string `json:"version,omitempty"`

	// ImagePullPolicy of the dgraph component.
	ImagePullPolicy *corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// Annotations of the component.
	Annotations map[string]string `json:"annotations,omitempty"`
}

// DgraphComponent represents a single member of either alpha or zero cluster.
type DgraphComponent struct {
	Name string `json:"name"`

	ID           string `json:"id"`
	ComponentURL string `json:"componentURL"`
	Healthy      bool   `json:"health"`
}

// +k8s:openapi-gen=true
// AlphaConfig is the configuration for dgraph alpha component.
type AlphaConfig struct {
	DgraphConfig `json:",inline"`
}

// +k8s:openapi-gen=true
// ZeroConfig is the configuration of dgraph zero component.
type ZeroConfig struct {
	DgraphConfig `json:",inline"`

	// ShardReplicaCount is the max number of replicas per data shard.
	ShardReplicaCount int32 `json:"shardReplicaCount,omitempty"`
}

// +k8s:openapi-gen=true
// DgraphConfig is the common configuration for dgraph components.
type DgraphConfig struct {
	// URL of the jaeger collector for dgraph alpha and zero components.
	JaegerCollector string `json:"jaegerCollector,omitempty"`
}
