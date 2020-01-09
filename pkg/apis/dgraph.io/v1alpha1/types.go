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
	"fmt"

	"github.com/dgraph-io/dgraph-operator/pkg/defaults"
	k8sutils "github.com/dgraph-io/dgraph-operator/pkg/k8s/utils"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterState represents the state of the cluster.
type ClusterState string

var (
	// ClusterStateCreating represents that the cluster is being created.
	ClusterStateCreating ClusterState = "creating"

	// ClusterStateRunning represents that the cluster is being udpated.
	ClusterStateRunning ClusterState = "running"

	// ClusterStateUpdating represents that the cluster is being udpated.
	ClusterStateUpdating ClusterState = "updating"
)

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

func (dc *DgraphCluster) internalComponentOverride(dcs *DgraphComponentSpec) {
	if dcs.Version == "" {
		dcs.Version = dc.Spec.Version
	}

	if dcs.ServiceType == "" {
		dcs.ServiceType = dc.Spec.ServiceType
	}

	if dcs.Resources == nil {
		dcs.Resources = dc.Spec.Resources.DeepCopy()
	}

	if dcs.ImagePullPolicy == nil {
		dcs.ImagePullPolicy = dc.Spec.ImagePullPolicy
	}
}

// ZeroClusterSpec returns cluster specification for dgraph zero component
// applying default values wherever necessery.
// It returns pointer to a copy of actual ZeroClusterSpec object in the
// DgraphCluster object.
func (dc *DgraphCluster) ZeroClusterSpec() *ZeroClusterSpec {
	zc := dc.Spec.ZeroCluster.DeepCopy()
	dc.internalComponentOverride(&zc.DgraphComponentSpec)

	return zc
}

// AlphaClusterSpec returns cluster specification for dgraph Alpha component
// applying default values wherever necessery.
// It returns pointer to a copy of actual AlphaClusterSpec object in the
// DgraphCluster object.
func (dc *DgraphCluster) AlphaClusterSpec() *AlphaClusterSpec {
	ac := dc.Spec.AlphaCluster.DeepCopy()
	dc.internalComponentOverride(&ac.DgraphComponentSpec)

	return ac
}

// RatelClusterSpec returns cluster specification for dgraph Ratel component
// applying default values wherever necessery.
// It returns pointer to a copy of actual RatelClusterSpec object in the
// DgraphCluster object.
func (dc *DgraphCluster) RatelClusterSpec() *RatelSpec {
	ac := dc.Spec.Ratel.DeepCopy()
	dc.internalComponentOverride(&ac.DgraphComponentSpec)

	return ac
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
	AlphaCluster *AlphaClusterSpec `json:"alpha"`

	// Cluster specification for dgraph zero components.
	ZeroCluster *ZeroClusterSpec `json:"zero"`

	// Specification for dgraph ratel component for providing UI.
	Ratel *RatelSpec `json:"ratel,omitempty"`

	// Below variables are more or less same
	// as that of DgraphComponentSpec but are cluster level, they can be overridden
	// inside individual component configuration.

	// Base image to use for dgraph cluster individual components, this can be overridden
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
	Resources *corev1.ResourceRequirements `json:"resources,omiempty"`
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
func (dc *DgraphClusterSpec) GetClusterID() string {
	return dc.ClusterID
}

// DgraphClusterStatus represents the status of a DgraphCluster.
type DgraphClusterStatus struct {
	// ClusterID is the ID of the dgraph cluster deployed.
	ClusterID string `json:"clusterID"`

	State ClusterState `json:"state"`

	// Status of individual dgraph components like alpha, zero and ratel.
	AlphaCluster AlphaClusterStatus `json:"alpha,omitempty"`
	ZeroCluster  ZeroClusterStatus  `json:"zero,omitempty"`
	Ratel        RatelStatus        `json:"ratel,omitempty"`
}

// +k8s:openapi-gen=true
// AlphaClusterSpec is the specification of the dgraph alpha cluster.
type AlphaClusterSpec struct {
	DgraphComponentSpec `json:",inline"`

	// Storage is the configuration for persistent storage for dgraph component.
	PersistentStorage *ComponentPersistentStorage `json:"persistentStorage,omitempty"`

	// Number of replicas to run in the cluster.
	Replicas int32 `json:"replicas"`

	// Config is the configuration of the dgraph component.
	Config *AlphaConfig `json:"config,omitempty"`
}

// LruMB returns the LRU MB configuration for dgraph alpha.
func (acs *AlphaClusterSpec) LruMB() int32 {
	if acs.Config == nil {
		return defaults.LruMBValue
	}
	return acs.Config.LruMB
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

	// PersistentStorage is the configuration for persistent storage for dgraph component.
	PersistentStorage *ComponentPersistentStorage `json:"persistentStorage,omitempty"`

	// Number of replicas to run in the cluster.
	Replicas int32 `json:"replicas"`

	// Config is the configuration of the dgraph zero.
	Config *ZeroConfig `json:"config,omitempty"`
}

// ShardReplicaCount returns the zero replica count to be used for deployment of the dgraph component.
func (zcs *ZeroClusterSpec) ShardReplicaCount() int32 {
	if zcs.Config == nil {
		return zcs.Replicas
	}
	return zcs.Config.ShardReplicaCount
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
	// Deployment is the status of stateful set associated with the specified
	// ratel cluster.
	Deployment *apps.DeploymentStatus `json:"deployment,omitempty"`

	// Members is the map of members in the zero cluster.
	Members map[string]DgraphComponent `json:"members,omitempty"`
}

// +k8s:openapi-gen=true
// DgraphComponentSpec is the common configuration values shared among different
// dgraph components.
type DgraphComponentSpec struct {
	// Resource requirements of the components.
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

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

// Image returns the image to be used for deployment of the dgraph component.
func (dcs *DgraphComponentSpec) Image() string {
	return fmt.Sprintf("%s:%s", dcs.BaseImage, dcs.Version)
}

// PodImagePullPolicy returns the image pull policy to be used for deployment of the dgraph component.
func (dcs *DgraphComponentSpec) PodImagePullPolicy() corev1.PullPolicy {
	return *dcs.ImagePullPolicy
}

// ResourceRequirements returns the resource requirements to be used for deployment of the dgraph component.
func (dcs *DgraphComponentSpec) ResourceRequirements() corev1.ResourceRequirements {
	if dcs.Resources == nil {
		return corev1.ResourceRequirements{}
	}

	return *dcs.Resources.DeepCopy()
}

// DgraphComponent represents a single member of either alpha or zero cluster.
type DgraphComponent struct {
	Name string `json:"name"`

	ID           string `json:"id"`
	ComponentURL string `json:"componentURL"`
	Healthy      bool   `json:"health"`
}

// ComponentPersistentStorage is the common type for storing configuration for
// persistent storage to associate with the dgraph component.
type ComponentPersistentStorage struct {
	// StorageClassName is the name of the storage class to use for the
	// persistent volumes for the dgraph component.
	StorageClassName string `json:"storageClassName,omitempty"`

	// Resource requirements for dgraph persistent storage.
	Requests corev1.ResourceList `json:"requests,omitempty"`
}

func (cps *ComponentPersistentStorage) StorageRequest() corev1.ResourceList {
	if cps.Requests == nil {
		return corev1.ResourceList{}
	}

	return cps.Requests.DeepCopy()
}

// +k8s:openapi-gen=true
// AlphaConfig is the configuration for dgraph alpha component.
type AlphaConfig struct {
	DgraphConfig `json:",inline"`

	// LruMB is the value of lrumb flag for dgraph alpha.
	LruMB int32 `json:"lruMB,omitempty"`
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
