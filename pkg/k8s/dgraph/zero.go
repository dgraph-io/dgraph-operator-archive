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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// DefaultZeroLabels returns a map representing labels associated with dgraph zero component.
func DefaultZeroLabels(instanceName string) map[string]string {
	zeroLabels := labels.NewLabelSet().
		Instance(instanceName).
		Component(defaults.ZeroMemberName).
		ManagedBy(defaults.DgraphOperatorName)

	return zeroLabels
}

// NewZeroService constructs a K8s service object for dgraph zero from the provided DgraphCluster
// configuration.
func NewZeroService(dc *v1alpha1.DgraphCluster) *corev1.Service {
	ns := dc.GetNamespace()
	name := dc.GetName()
	clusterID := dc.Spec.GetClusterID()

	serviceName := utils.DgraphZeroMemberName(clusterID, name)
	zeroLabels := DefaultZeroLabels(serviceName)

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            serviceName,
			Namespace:       ns,
			Labels:          zeroLabels,
			OwnerReferences: []metav1.OwnerReference{dc.AsOwnerReference()},
		},
		Spec: corev1.ServiceSpec{
			Type: dc.Spec.ZeroServiceType(),
			Ports: []corev1.ServicePort{
				{
					Name:       defaults.ZeroGRPCPortName,
					Port:       defaults.ZeroGRPCPort,
					TargetPort: intstr.FromInt(int(defaults.ZeroGRPCPort)),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       defaults.ZeroHTTPPortName,
					Port:       defaults.ZeroHTTPPort,
					TargetPort: intstr.FromInt(int(defaults.ZeroHTTPPort)),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: zeroLabels,
		},
	}
}
