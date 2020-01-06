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

package dgraphcluster

import (
	"reflect"

	dgraphio "github.com/dgraph-io/dgraph-operator/pkg/apis/dgraph.io/v1alpha1"
)

// UpdateDgraphCluster function handles the udpate of dgraph cluster object.
func (dc *Controller) UpdateDgraphCluster(dcObj *dgraphio.DgraphCluster) error {
	oldStatus := dcObj.Status.DeepCopy()

	// During update we relay the logic of update to the respective managers which
	// are the managers for individual top level resource as understood by DgraphCluster
	// This is because they may have different strategies for update being different in
	// terms of usage.
	//
	// These top level resources includes
	// 1. Alpha
	// 2. Zero
	// 3. Ratel
	//
	// Each manager individually syncs the underlying kubernetes resources it manages.
	for _, manager := range dc.managers {
		if err := manager.Sync(dcObj); err != nil {
			return err
		}
	}

	// Check if the status is same as the old status or not, if not then udpate the
	// status of the DgraphCluster object.
	if !reflect.DeepEqual(dcObj.Status, oldStatus) {
		if err := dc.UpdateDgraphClusterStatus(dcObj, &dcObj.Status); err != nil {
			return err
		}
	}

	return nil
}

// UpdateDgraphClusterStatus updates the status of the DgraphCluster object represented by dcObj
// with the status represented in dcStatus.
func (dc *Controller) UpdateDgraphClusterStatus(
	dcObj *dgraphio.DgraphCluster,
	dcStatus *dgraphio.DgraphClusterStatus) error {
	return nil
}
