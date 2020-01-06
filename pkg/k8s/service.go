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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

// CreateNewService creates a new Kubernetes service for the provided
// service object.
func CreateNewService(k8sClient kubernetes.Interface, namespace string, svc *corev1.Service) error {
	_, err := k8sClient.CoreV1().
		Services(namespace).
		Create(svc)
	return err
}

// UpdateService updates the service in the kubernetes cluster.
func UpdateService(k8sClient kubernetes.Interface, namespace string, svc *corev1.Service) (*corev1.Service, error) {
	var updatedSvc *corev1.Service
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		var updateErr error
		updatedSvc, updateErr = k8sClient.CoreV1().
			Services(namespace).
			Update(svc)

		return updateErr
	})

	return updatedSvc, err
}

// DeleteService deletes a kubernetes service from the cluster.
func DeleteService(k8sClient kubernetes.Interface, namespace string, svc *corev1.Service) error {
	return k8sClient.CoreV1().
		Services(namespace).
		Delete(svc.Name, nil)
}
