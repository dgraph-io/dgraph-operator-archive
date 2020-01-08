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
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

// CreateNewDeployment creates a new Kubernetes Deployment for the provided
// Deployment object.
func CreateNewDeployment(k8sClient kubernetes.Interface, namespace string, svc *appsv1.Deployment) error {
	_, err := k8sClient.AppsV1().
		Deployments(namespace).
		Create(svc)
	return err
}

// UpdateDeployment updates the Deployment in the kubernetes cluster.
func UpdateDeployment(k8sClient kubernetes.Interface, namespace string,
	svc *appsv1.Deployment) (*appsv1.Deployment, error) {
	var updatedDeployment *appsv1.Deployment
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		var updateErr error
		updatedDeployment, updateErr = k8sClient.AppsV1().
			Deployments(namespace).
			Update(svc)

		return updateErr
	})

	return updatedDeployment, err
}

// DeleteDeployment deletes a kubernetes Deployment from the cluster.
func DeleteDeployment(k8sClient kubernetes.Interface, namespace string, svc *appsv1.Deployment) error {
	return k8sClient.AppsV1().
		Deployments(namespace).
		Delete(svc.Name, nil)
}
