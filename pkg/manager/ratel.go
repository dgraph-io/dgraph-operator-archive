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

// RatelManager manages Ratel members in a dgraph cluster. It's main function is to sync
// the Ratels in the clsuter with the required configuration in the DgraphCluster object.
type RatelManager struct {
	k8sClient kubernetes.Interface

	podLister        klisters.PodLister
	svcLister        klisters.ServiceLister
	deploymentLister v1.DeploymentLister
}

// NewRatelManager creates a new manager for dgraph Ratel components.
func NewRatelManager(
	k8sClient kubernetes.Interface,
	podLister klisters.PodLister,
	svcLister klisters.ServiceLister,
	deploymentLister v1.DeploymentLister,
) *RatelManager {
	return &RatelManager{
		k8sClient,
		podLister,
		svcLister,
		deploymentLister,
	}
}

func (rm *RatelManager) Sync(dc *v1alpha1.DgraphCluster) error {
	glog.Info("syncing dgraph ratel components.")
	if dc.Spec.Ratel == nil {
		glog.Info("no configuration for ratel provided, skipping")
		return nil
	}
	if err := rm.syncRatelServiceWithDgraphCluster(dc); err != nil {
		return err
	}

	return rm.syncRatelDeploymentWithDgraphCluster(dc)
}

func (rm *RatelManager) syncRatelServiceWithDgraphCluster(dc *v1alpha1.DgraphCluster) error {
	ns := dc.GetNamespace()
	svc := dgraphk8s.NewRatelService(dc)
	serviceName := svc.GetName()
	glog.Infof("syncing dgraph ratel service(%s) with dgraph cluster specification", serviceName)

	oldSVC, err := rm.svcLister.Services(ns).Get(serviceName)
	if kerrors.IsNotFound(err) {
		// Existing service not found create a new one.
		glog.Infof("creating new service for dgraph ratel: %s", svc.GetName())
		return k8s.CreateNewService(rm.k8sClient, ns, svc)
	}
	if err != nil {
		return err
	}

	// If the old service and new service spec is same don't change anything.
	// else update the service spec as mentioned in the updated specification.
	svc.Spec.ClusterIP = oldSVC.Spec.ClusterIP

	if !apiequality.Semantic.DeepDerivative(svc.Spec, oldSVC.Spec) {
		updateSVC := *oldSVC
		updateSVC.Spec = svc.Spec
		glog.Info("updating service for dgraph ratel")
		if _, err = k8s.UpdateService(rm.k8sClient, ns, &updateSVC); err != nil {
			return err
		}
	}

	return nil
}

func (rm *RatelManager) syncRatelDeploymentWithDgraphCluster(dc *v1alpha1.DgraphCluster) error {
	ns := dc.GetNamespace()
	deployment := dgraphk8s.NewRatelDeployment(dc)
	deploymentName := deployment.GetName()
	glog.Infof("syncing dgraph ratel Deployment(%s) with dgraph cluster specification", deploymentName)

	oldDeployment, err := rm.deploymentLister.Deployments(ns).Get(deploymentName)
	if kerrors.IsNotFound(err) {
		// Existing Deployment not found create a new one.
		glog.Infof("creating new Deployment for dgraph ratel: %s", deployment.GetName())
		return k8s.CreateNewDeployment(rm.k8sClient, ns, deployment)
	}
	if err != nil {
		return err
	}

	if !apiequality.Semantic.DeepDerivative(deployment.Spec, oldDeployment.Spec) {
		updateDeployment := *oldDeployment
		updateDeployment.Spec = deployment.Spec
		glog.Info("updating Deployment for dgraph ratel")
		if _, err = k8s.UpdateDeployment(rm.k8sClient, ns, &updateDeployment); err != nil {
			return err
		}
	}

	return nil
}
