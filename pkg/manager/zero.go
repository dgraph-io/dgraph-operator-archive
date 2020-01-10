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

package manager

import (
	"github.com/dgraph-io/dgraph-operator/pkg/apis/dgraph.io/v1alpha1"
	"github.com/dgraph-io/dgraph-operator/pkg/k8s"
	dgraphk8s "github.com/dgraph-io/dgraph-operator/pkg/k8s/dgraph"
	"github.com/dgraph-io/dgraph-operator/pkg/utils"

	"github.com/golang/glog"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/listers/apps/v1"
	klisters "k8s.io/client-go/listers/core/v1"
)

// ZeroManager manages Zero members in a dgraph cluster. It's main function is to sync
// dgraph Zeros in the clsuter with the required configuration in the DgraphCluster object.
type ZeroManager struct {
	// k8sClient is the kubernetes client to interact with the Kubernetes API server.
	k8sClient kubernetes.Interface

	// These represents listers which are used by managers in order to list
	// actual resources in Kubernetes cluster.
	// Zero only deals with below mentioned Kubernetes resources for running the
	// cluster:
	// 1. Pods
	// 2. Services: There are two types of services associated with dgraph zero
	//      - Headless Service
	//      - Kubernetes Service as specified by user in serviceType
	// 3. StatefulSet: StatefulSet is the actual Kubernetes abstraction under which
	//      the zero instances run.
	podLister         klisters.PodLister
	svcLister         klisters.ServiceLister
	statefulSetLister v1.StatefulSetLister
}

// NewZeroManager creates a new manager for dgraph zero components
// in the clsuter.
func NewZeroManager(
	k8sClient kubernetes.Interface,
	podLister klisters.PodLister,
	svcLister klisters.ServiceLister,
	statefulSetLister v1.StatefulSetLister,
) *ZeroManager {
	return &ZeroManager{
		k8sClient,
		podLister,
		svcLister,
		statefulSetLister,
	}
}

// Sync syncs the actual Kubernetes resources with the provided configuration
// of DgraphCluster using operator defined custom resources.
func (zm *ZeroManager) Sync(dc *v1alpha1.DgraphCluster) error {
	glog.Infof("zero-manager: syncing dgraph zero components for cluster: %s", dc.GetName())

	if err := zm.syncZeroServiceWithDgraphCluster(dc); err != nil {
		return err
	}

	return zm.syncZeroStatefulSetWithDgraphCluster(dc)
}

// syncZeroServiceWithDgraphCluster syncs the dgraph zero service with the DgraphCluster
// specification provided.
//
// Here we create a new service **type** for dgraph zero and compare it with existing service
// associated with dgraph zero cluster. If there is a difference between the two
// we update the service to match the configuration in latest DgraphCluster object.
//
// This function handle the creation both kind of services for dgraph zero cluster which includes
// 1. Service(ClusterIP, NodePort or LoadBalancer)
// 2. Headless Service - ClusterIP with ClusterIP None
func (zm *ZeroManager) syncZeroServiceWithDgraphCluster(dc *v1alpha1.DgraphCluster) error {
	ns := dc.GetNamespace()

	// create a new service type for zero using the dgraphcluster configuration
	// we have.
	svc := dgraphk8s.NewZeroService(dc)
	serviceName := svc.GetName()
	glog.Infof("zero-manager: syncing dgraph zero service(%s) with "+
		"dgraph cluster specification", serviceName)

	// check if there is an already existing service with the provided name.
	oldSVC, err := zm.svcLister.Services(ns).Get(serviceName)
	if kerrors.IsNotFound(err) {
		// Existing service not found create a new one.
		glog.Infof("zero-manager: creating new service for dgraph zero: %s", svc.GetName())
		return k8s.CreateNewService(zm.k8sClient, ns, svc)
	}
	if err != nil {
		return err
	}

	// If the old service and new service spec is same don't change anything.
	// else update the service spec as mentioned in the updated specification.
	svc.Spec.ClusterIP = oldSVC.Spec.ClusterIP

	// We don't use DeepEquals here to not compare the default values that
	// kubernetes might have introduced to the given type.
	// TODO: Improve this.
	if !apiequality.Semantic.DeepDerivative(svc.Spec, oldSVC.Spec) {
		updateSVC := *oldSVC
		updateSVC.Spec = svc.Spec
		glog.Info("zero-manager: updating service for dgraph zero")
		if _, err = k8s.UpdateService(zm.k8sClient, ns, &updateSVC); err != nil {
			return err
		}
	} else {
		glog.Info("zero-manager: no change found in dgraph zero service")
	}

	glog.Info("zero-manager: syncing dgraph zero headless service with dgraph " +
		" cluster specification")
	headlessSVC := dgraphk8s.NewZeroHeadlessService(dc)
	oldHeadlessSVC, err := zm.svcLister.Services(ns).Get(headlessSVC.GetObjectMeta().GetName())
	if kerrors.IsNotFound(err) {
		// Existing service not found create a new one.
		glog.Info("zero-manager: creating new headless service for dgraph zero")
		return k8s.CreateNewService(zm.k8sClient, ns, headlessSVC)
	}
	if err != nil {
		return err
	}

	// If the old service and new service spec is same don't change anything.
	if apiequality.Semantic.DeepDerivative(headlessSVC.Spec, oldHeadlessSVC.Spec) {
		glog.Info("zero-manager: no change found in dgraph zero headless service")
		return nil
	}

	headlessSVCUpdate := *oldHeadlessSVC
	headlessSVCUpdate.Spec = headlessSVC.Spec
	glog.Info("zero-manager: udpating headless service for dgraph zero")
	_, err = k8s.UpdateService(zm.k8sClient, ns, &headlessSVCUpdate)

	return err
}

// syncZeroStatefulSetWithDgraphCluster syncs the dgraph zero service with the DgraphCluster
// specification provided.
func (zm *ZeroManager) syncZeroStatefulSetWithDgraphCluster(dc *v1alpha1.DgraphCluster) error {
	glog.Info("zero-manager: syncing dgraph zero stateful set with dgraph cluster specification")
	ns := dc.GetNamespace()

	zeroStatefulSet := dgraphk8s.NewZeroStatefulSet(dc)
	zeroStatefulSetOld, err := zm.statefulSetLister.StatefulSets(ns).
		Get(utils.DgraphZeroMemberName(dc.Spec.GetClusterID(), dc.GetObjectMeta().GetName()))
	if kerrors.IsNotFound(err) {
		glog.Info("zero-manager: creating new stateful set for zero for DgraphCluster spec")
		return k8s.CreateNewStatefulSet(zm.k8sClient, ns, zeroStatefulSet)
	}
	if err != nil {
		return err
	}

	// If the old service and new service spec is same don't change anything.
	if apiequality.Semantic.DeepDerivative(zeroStatefulSet.Spec, zeroStatefulSetOld.Spec) {
		glog.Info("zero-manager: no change found for dgraph zero stateful set spec")
		return nil
	}

	statefulSetUpdate := *zeroStatefulSetOld
	statefulSetUpdate.Spec = zeroStatefulSet.Spec
	glog.Infof("udpating underlying stateful set for dgraph zero")
	_, err = k8s.UpdateStatefulSet(zm.k8sClient, ns, &statefulSetUpdate)

	return err
}
