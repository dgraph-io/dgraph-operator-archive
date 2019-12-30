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

package main

import (
	dgraphio "github.com/dgraph-io/dgraph-operator/pkg/apis/dgraph.io/v1alpha1"
	"github.com/dgraph-io/dgraph-operator/pkg/k8s"
	"github.com/golang/glog"
)

// RunOperator runs an operator with the provided configuration.
func RunOperator() {
	apiExtClient, err := k8s.APIExtClient()
	if err != nil {
		glog.Errorf("error while configuring apiextension client: %s", err)
		return
	}

	if err = dgraphio.CreateCustomResourceDefinitions(apiExtClient); err != nil {
		glog.Errorf("error while creating operator CRDs: %s", err)
		return
	}
}
