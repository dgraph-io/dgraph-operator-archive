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

// CreateNewStatefulSet creates a new Kubernetes StatefulSet for the provided
// StatefulSet object.
func CreateNewStatefulSet(k8sClient kubernetes.Interface, namespace string, svc *appsv1.StatefulSet) error {
	_, err := k8sClient.AppsV1().
		StatefulSets(namespace).
		Create(svc)
	return err
}

// UpdateStatefulSet updates the StatefulSet in the kubernetes cluster.
func UpdateStatefulSet(k8sClient kubernetes.Interface, namespace string,
	svc *appsv1.StatefulSet) (*appsv1.StatefulSet, error) {
	var updatedStatefulSet *appsv1.StatefulSet
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		var updateErr error
		updatedStatefulSet, updateErr = k8sClient.AppsV1().
			StatefulSets(namespace).
			Update(svc)

		return updateErr
	})

	return updatedStatefulSet, err
}

// DeleteStatefulSet deletes a kubernetes StatefulSet from the cluster.
func DeleteStatefulSet(k8sClient kubernetes.Interface, namespace string, svc *appsv1.StatefulSet) error {
	return k8sClient.AppsV1().
		StatefulSets(namespace).
		Delete(svc.Name, nil)
}
