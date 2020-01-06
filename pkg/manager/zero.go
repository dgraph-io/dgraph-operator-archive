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
func (zm *ZeroManager) syncZeroServiceWithDgraphCluster(dc *v1alpha1.DgraphCluster) error {
	glog.Info("syncing dgraph zero service with dgraph cluster specification")
	ns := dc.GetNamespace()
	name := dc.GetName()

	svc := dgraphk8s.NewZeroService(dc)
	oldSVC, err := zm.svcLister.Services(ns).Get(name)
	if kerrors.IsNotFound(err) {
		// Existing service not found create a new one.
		glog.Info("creating new service for dgraph zero")
		return k8s.CreateNewService(zm.k8sClient, ns, svc)
	}
	if err != nil {
		return err
	}

	// If the old service and new service spec is same don't change anything.
	if apiequality.Semantic.DeepEqual(svc.Spec, oldSVC.Spec) {
		return nil
	}

	updateSVC := *oldSVC
	updateSVC.Spec = svc.Spec
	glog.Info("udpating service for dgraph zero")
	_, err = k8s.UpdateService(zm.k8sClient, ns, &updateSVC)
	return err
}

// syncZeroStatefulSetWithDgraphCluster syncs the dgraph zero service with the DgraphCluster
// specification provided.
func (zm *ZeroManager) syncZeroStatefulSetWithDgraphCluster(dc *v1alpha1.DgraphCluster) error {
	glog.Info("syncing dgraph zero stateful set with dgraph cluster specification")
	return nil
}
