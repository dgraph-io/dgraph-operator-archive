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
	"github.com/golang/glog"

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

func (am *AlphaManager) Sync(dc *v1alpha1.DgraphCluster) error {
	glog.Info("syncing dgraph alpha components.")

	return nil
}
