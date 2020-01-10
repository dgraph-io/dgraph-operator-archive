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
	"github.com/dgraph-io/dgraph-operator/pkg/defaults"
	"github.com/dgraph-io/dgraph-operator/pkg/labels"
	"github.com/dgraph-io/dgraph-operator/pkg/utils"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// DefaultRatelLabels returns a map representing labels associated with dgraph Ratel component.
func DefaultRatelLabels(instanceName string) map[string]string {
	ratelLabels := labels.NewLabelSet().
		Instance(instanceName).
		Component(defaults.RatelMemberName).
		ManagedBy(defaults.DgraphOperatorName)

	return ratelLabels
}

// NewRatelService constructs a K8s service object for dgraph Ratel from the provided DgraphCluster
// configuration.
func NewRatelService(dc *v1alpha1.DgraphCluster) *corev1.Service {
	ns := dc.GetNamespace()
	name := dc.GetName()
	clusterID := dc.Spec.GetClusterID()

	serviceName := utils.DgraphRatelMemberName(clusterID, name)
	ratelLabels := DefaultRatelLabels(serviceName)

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            serviceName,
			Namespace:       ns,
			Labels:          ratelLabels,
			OwnerReferences: []metav1.OwnerReference{dc.AsOwnerReference()},
		},
		Spec: corev1.ServiceSpec{
			Type: dc.Spec.RatelServiceType(),
			Ports: []corev1.ServicePort{
				{
					Name:       defaults.RatelPortName,
					Port:       defaults.RatelPort,
					TargetPort: intstr.FromInt(int(defaults.RatelPort)),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: ratelLabels,
		},
	}
}

// NewRatelDeployment constructs a K8s Deployment object for dgraph Ratel from
// the provided DgraphCluster configuration.
func NewRatelDeployment(dc *v1alpha1.DgraphCluster) *appsv1.Deployment {
	ns := dc.GetNamespace()
	name := dc.GetName()
	clusterID := dc.Spec.GetClusterID()

	deploymentName := utils.DgraphRatelMemberName(clusterID, name)
	ratelLabels := DefaultRatelLabels(deploymentName)
	replicas := dc.Spec.Ratel.Replicas

	// POD spec for the deployment.
	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:            deploymentName,
				Image:           dc.RatelClusterSpec().Image(),
				ImagePullPolicy: dc.RatelClusterSpec().PodImagePullPolicy(),
				Command: []string{
					"dgraph-ratel",
				},
				Ports: []corev1.ContainerPort{
					{
						Name:          defaults.RatelPortName,
						ContainerPort: defaults.RatelPort,
						Protocol:      corev1.ProtocolTCP,
					},
				},
				// Environment variables to be used within the container.
				Env: []corev1.EnvVar{
					{
						Name: "POD_NAMESPACE",
						ValueFrom: &corev1.EnvVarSource{
							FieldRef: &corev1.ObjectFieldSelector{
								FieldPath: "metadata.namespace",
							},
						},
					},
				},
				Resources: dc.RatelClusterSpec().ResourceRequirements(),
			},
		},
		RestartPolicy: corev1.RestartPolicyAlways,
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            deploymentName,
			Namespace:       ns,
			Labels:          ratelLabels,
			OwnerReferences: []metav1.OwnerReference{dc.AsOwnerReference()},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ratelLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ratelLabels,
				},
				Spec: podSpec,
			},
		},
	}
}
