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
	"github.com/dgraph-io/dgraph-operator/pkg/controller"
	"github.com/dgraph-io/dgraph-operator/pkg/k8s"
	"github.com/dgraph-io/dgraph-operator/pkg/option"
	"github.com/golang/glog"
)

// RunOperator bootstraps the configuration for operator and then initiate controller
// run.
func RunOperator() {
	// update k8s version using the client.
	client, err := k8s.Client()
	if err != nil {
		glog.Fatalf("error while creating kubernetes client: %s", err)
	}
	if err := k8s.UpdateVersion(client); err != nil {
		glog.Fatalf("error updating kubernetes server version: %s", err)
	}

	if !option.OperatorConfig.SkipCRDCreation && k8s.CanUseAPIExtV1() {
		apiExtClient, err := k8s.APIExtClient()
		if err != nil {
			glog.Fatalf("error while configuring apiextension client: %s", err)
		}

		// Create the custom resource definition for dgraph-operator if they don't exist.
		if err = dgraphio.CreateCustomResourceDefinitions(apiExtClient); err != nil {
			glog.Fatalf("error while creating operator CRDs: %s", err)
		}
	} else {
		if !k8s.CanUseAPIExtV1() {
			glog.Warningf("k8s version %s does not support CRD creation using apiextension", k8s.Version())
		}
		glog.Info("skipping automatic crd creation for operator")
	}

	cm := controller.MustNewControllerManager()
	// Run operator controllers to watch for the resources created in kubernetes context
	// and perform some action based on that.
	if err := cm.RunOperatorControllers(); err != nil {
		glog.Fatalf("error while setting up controller for operator: %s", err)
	}
}
