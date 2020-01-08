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
// the Zeros in the clsuter with the required configuration in the DgraphCluster object.
type ZeroManager struct {
	k8sClient kubernetes.Interface

	podLister         klisters.PodLister
	svcLister         klisters.ServiceLister
	statefulSetLister v1.StatefulSetLister
}

// NewZeroManager creates a new manager for dgraph Zero components.
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

func (zm *ZeroManager) Sync(dc *v1alpha1.DgraphCluster) error {
	glog.Info("syncing dgraph zero components.")

	if err := zm.syncZeroServiceWithDgraphCluster(dc); err != nil {
		return err
	}

	return zm.syncZeroStatefulSetWithDgraphCluster(dc)
}

// syncZeroServiceWithDgraphCluster syncs the dgraph zero service with the DgraphCluster
// specification provided.
//
// Here we create a new service type for dgraph zero and compare it with existing service
// associated with dgraph zero cluster. If there is a difference between the two
// we update the service to match the configuration in latest DgraphCluster object.
//
// This function handle the creation both kind of services for dgraph zero cluster which includes
// 1. Service(ClusterIP, NodePort or LoadBalancer)
// 2. Headless Service - ClusterIP with ClusterIP None
func (zm *ZeroManager) syncZeroServiceWithDgraphCluster(dc *v1alpha1.DgraphCluster) error {
	ns := dc.GetNamespace()
	svc := dgraphk8s.NewZeroService(dc)
	serviceName := svc.GetName()
	glog.Infof("syncing dgraph zero service(%s) with dgraph cluster specification", serviceName)

	oldSVC, err := zm.svcLister.Services(ns).Get(serviceName)
	if kerrors.IsNotFound(err) {
		// Existing service not found create a new one.
		glog.Infof("creating new service for dgraph zero: %s", svc.GetName())
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
		glog.Info("updating service for dgraph zero")
		if _, err = k8s.UpdateService(zm.k8sClient, ns, &updateSVC); err != nil {
			return err
		}
	}

	glog.Info("syncing dgraph zero headless service with dgraph cluster specification")
	headlessSVC := dgraphk8s.NewZeroHeadlessService(dc)
	oldHeadlessSVC, err := zm.svcLister.Services(ns).Get(headlessSVC.GetObjectMeta().GetName())
	if kerrors.IsNotFound(err) {
		// Existing service not found create a new one.
		glog.Info("creating new headless service for dgraph zero")
		return k8s.CreateNewService(zm.k8sClient, ns, headlessSVC)
	}
	if err != nil {
		return err
	}

	// If the old service and new service spec is same don't change anything.
	if apiequality.Semantic.DeepDerivative(headlessSVC.Spec, oldHeadlessSVC.Spec) {
		return nil
	}

	headlessSVCUpdate := *oldHeadlessSVC
	headlessSVCUpdate.Spec = headlessSVC.Spec
	glog.Info("udpating headless service for dgraph zero")
	_, err = k8s.UpdateService(zm.k8sClient, ns, &headlessSVCUpdate)

	return err
}

// syncZeroStatefulSetWithDgraphCluster syncs the dgraph zero service with the DgraphCluster
// specification provided.
func (zm *ZeroManager) syncZeroStatefulSetWithDgraphCluster(dc *v1alpha1.DgraphCluster) error {
	glog.Info("syncing dgraph zero stateful set with dgraph cluster specification")
	ns := dc.GetNamespace()

	zeroStatefulSet := dgraphk8s.NewZeroStatefulSet(dc)
	zeroStatefulSetOld, err := zm.statefulSetLister.StatefulSets(ns).
		Get(utils.DgraphZeroMemberName(dc.Spec.GetClusterID(), dc.GetObjectMeta().GetName()))
	if kerrors.IsNotFound(err) {
		glog.Info("creating new stateful set for zero corresponding to DgraphCluster configuration spec")
		return k8s.CreateNewStatefulSet(zm.k8sClient, ns, zeroStatefulSet)
	}
	if err != nil {
		return err
	}

	// If the old service and new service spec is same don't change anything.
	if apiequality.Semantic.DeepDerivative(zeroStatefulSet.Spec.Template, zeroStatefulSetOld.Spec.Template) {
		return nil
	}

	statefulSetUpdate := *zeroStatefulSetOld
	statefulSetUpdate.Spec = zeroStatefulSet.Spec
	glog.Infof("udpating underlying stateful set for dgraph zero")
	_, err = k8s.UpdateStatefulSet(zm.k8sClient, ns, &statefulSetUpdate)

	return err
}
