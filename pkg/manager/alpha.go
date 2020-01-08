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

// AlphaManager manages alpha members in a dgraph cluster. It's main function is to sync
// the alphas in the clsuter with the required configuration in the DgraphCluster object.
type AlphaManager struct {
	k8sClient kubernetes.Interface

	podLister         klisters.PodLister
	svcLister         klisters.ServiceLister
	statefulSetLister v1.StatefulSetLister
}

// NewAlphaManager creates a new manager for dgraph alpha components.
func NewAlphaManager(
	k8sClient kubernetes.Interface,
	podLister klisters.PodLister,
	svcLister klisters.ServiceLister,
	statefulSetLister v1.StatefulSetLister,
) *AlphaManager {
	return &AlphaManager{
		k8sClient,
		podLister,
		svcLister,
		statefulSetLister,
	}
}

// Sync syncs the alpha cluster for the provided DgraphCluster configuration.
func (am *AlphaManager) Sync(dc *v1alpha1.DgraphCluster) error {
	if err := am.syncAlphaServiceWithDgraphCluster(dc); err != nil {
		return err
	}

	return am.syncAlphaStatefulSetWithDgraphCluster(dc)
}

// syncAlphaServiceWithDgraphCluster syncs the dgraph alpha service with the DgraphCluster
// specification provided.
//
// Here we create a new service type for dgraph Alpha and compare it with existing service
// associated with dgraph Alpha cluster. If there is a difference between the two
// we update the service to match the configuration in latest DgraphCluster object.
//
// This function handle the creation both kind of services for dgraph Alpha cluster which includes
// 1. Service(ClusterIP, NodePort or LoadBalancer)
// 2. Headless Service - ClusterIP with ClusterIP None
func (am *AlphaManager) syncAlphaServiceWithDgraphCluster(dc *v1alpha1.DgraphCluster) error {
	ns := dc.GetNamespace()
	svc := dgraphk8s.NewAlphaService(dc)
	serviceName := svc.GetName()
	glog.Infof("syncing dgraph alpha service(%s) with dgraph cluster specification", serviceName)

	oldSVC, err := am.svcLister.Services(ns).Get(serviceName)
	if kerrors.IsNotFound(err) {
		// Existing service not found create a new one.
		glog.Infof("creating new service for dgraph alpha: %s", svc.GetName())
		return k8s.CreateNewService(am.k8sClient, ns, svc)
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
		glog.Info("updating service for dgraph alpha")
		if _, err = k8s.UpdateService(am.k8sClient, ns, &updateSVC); err != nil {
			return err
		}
	}

	glog.Info("syncing dgraph alpha headless service with dgraph cluster specification")
	headlessSVC := dgraphk8s.NewAlphaHeadlessService(dc)
	oldHeadlessSVC, err := am.svcLister.Services(ns).Get(headlessSVC.GetObjectMeta().GetName())
	if kerrors.IsNotFound(err) {
		// Existing service not found create a new one.
		glog.Info("creating new headless service for dgraph alpha")
		return k8s.CreateNewService(am.k8sClient, ns, headlessSVC)
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
	glog.Info("udpating headless service for dgraph alpha")
	_, err = k8s.UpdateService(am.k8sClient, ns, &headlessSVCUpdate)

	return err
}

// syncAlphaStatefulSetWithDgraphCluster syncs the dgraph Alpha service with the DgraphCluster
// specification provided.
func (am *AlphaManager) syncAlphaStatefulSetWithDgraphCluster(dc *v1alpha1.DgraphCluster) error {
	glog.Info("syncing dgraph Alpha stateful set with dgraph cluster specification")
	ns := dc.GetNamespace()

	AlphaStatefulSet := dgraphk8s.NewAlphaStatefulSet(dc)
	AlphaStatefulSetOld, err := am.statefulSetLister.StatefulSets(ns).
		Get(utils.DgraphAlphaMemberName(dc.Spec.GetClusterID(), dc.GetObjectMeta().GetName()))
	if kerrors.IsNotFound(err) {
		glog.Info("creating new stateful set for alpha according to DgraphCluster configuration spec")
		return k8s.CreateNewStatefulSet(am.k8sClient, ns, AlphaStatefulSet)
	}
	if err != nil {
		return err
	}

	// If the old service and new service spec is same don't change anything.
	if apiequality.Semantic.DeepDerivative(AlphaStatefulSet.Spec.Template, AlphaStatefulSetOld.Spec.Template) {
		return nil
	}

	statefulSetUpdate := *AlphaStatefulSetOld
	statefulSetUpdate.Spec = AlphaStatefulSet.Spec
	glog.Infof("udpating underlying stateful set for dgraph alpha")
	_, err = k8s.UpdateStatefulSet(am.k8sClient, ns, &statefulSetUpdate)

	return err
}
