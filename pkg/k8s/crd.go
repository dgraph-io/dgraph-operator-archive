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

package k8s

import (
	"github.com/dgraph-io/dgraph-operator/pkg/defaults"

	"github.com/golang/glog"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// WaitForCRD waits for a kubernetes custom resource definition to be ready.
func WaitForCRD(crdName string) error {
	apiExtClient, err := APIExtClient()
	if err != nil {
		return err
	}

	// Wait for the CRD to be available
	glog.Info("Waiting for CRD (CustomResourceDefinition) to be available...")
	err = wait.Poll(defaults.CRDWaitPollInterval, defaults.K8SAPIServerRequestTimeout,
		func() (bool, error) {
			crd, err := apiExtClient.
				ApiextensionsV1().
				CustomResourceDefinitions().
				Get(crdName, metav1.GetOptions{})
			if err != nil {
				return false, err
			}
			for _, cond := range crd.Status.Conditions {
				switch cond.Type {
				case apiextv1.Established:
					if cond.Status == apiextv1.ConditionTrue {
						return true, err
					}
				case apiextv1.NamesAccepted:
					if cond.Status == apiextv1.ConditionFalse {
						glog.Errorf("name conflict for CRD: %s", crdName)
						return false, err
					}
				}
			}
			return false, err
		})

	return err
}
