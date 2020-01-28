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
	"fmt"

	"github.com/dgraph-io/dgraph-operator/pkg/apis/dgraph.io/v1alpha1"
	"github.com/dgraph-io/dgraph-operator/pkg/defaults"
	"github.com/dgraph-io/dgraph-operator/pkg/labels"
	"github.com/dgraph-io/dgraph-operator/pkg/utils"

	appsv1 "k8s.io/api/apps/v1"
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

// NewZeroHeadlessService constructs a K8s headless service object for dgraph zero from the provided
// DgraphCluster configuration.
func NewZeroHeadlessService(dc *v1alpha1.DgraphCluster) *corev1.Service {
	svc := NewZeroService(dc)

	// Service name for headless service is of the format
	// <clusterID>-<clusterName>-zero-headless
	svc.Name = fmt.Sprintf("%s%s%s", svc.Name, defaults.K8SDelimeter, defaults.HeadlessServiceSuffix)
	// Change spec for kubernetes headless service
	svc.Spec = corev1.ServiceSpec{
		ClusterIP:                "None",
		Ports:                    svc.Spec.Ports,
		Selector:                 svc.Spec.Selector,
		PublishNotReadyAddresses: true,
	}

	return svc
}

// NewZeroStatefulSet constructs a K8s stateful set object for dgraph zero from the
// provided DgraphCluster configuration.
func NewZeroStatefulSet(dc *v1alpha1.DgraphCluster) *appsv1.StatefulSet {
	ns := dc.GetNamespace()
	name := dc.GetName()
	clusterID := dc.Spec.GetClusterID()

	ssName := utils.DgraphZeroMemberName(clusterID, name)
	headlessServiceName := fmt.Sprintf("%s%s%s",
		ssName,
		defaults.K8SDelimeter,
		defaults.HeadlessServiceSuffix)
	storageClassName := dc.Spec.ZeroCluster.PersistentStorage.StorageClassName
	shardReplicaCount := dc.Spec.ZeroCluster.ShardReplicaCount()
	zeroLabels := DefaultZeroLabels(ssName)

	replicaCount := dc.Spec.ZeroCluster.Replicas
	partitionCount := dc.Spec.ZeroCluster.Replicas

	// nolint
	zeroRunCmd := fmt.Sprintf(`set -ex
[[ $(hostname) =~ -([0-9]+)$ ]] || exit 1
ordinal=${BASH_REMATCH[1]}
idx=$(($ordinal + 1))
if [[ $ordinal -eq 0 ]]; then
    exec dgraph zero --my=$(hostname -f):5080 --idx $idx --replicas %d
else
    exec dgraph zero --my=$(hostname -f):5080 --peer %s-0.%s.${POD_NAMESPACE}.svc.cluster.local:5080 \
        --idx $idx --replicas %d
fi`, shardReplicaCount, ssName, headlessServiceName, shardReplicaCount)

	podVolumeMounts := []corev1.VolumeMount{
		{
			Name:      ssName,
			MountPath: defaults.ZeroPersistentVolumeMountPath,
		},
	}

	podAffinity := &corev1.Affinity{
		PodAntiAffinity: &corev1.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
				{
					Weight: defaults.StatefulSetPodAntiAffinityWeight,
					PodAffinityTerm: corev1.PodAffinityTerm{
						TopologyKey: defaults.StatefulSetPodAntiAffinityKey,
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: labels.NewLabelSet().Component(defaults.ZeroMemberName),
						},
					},
				},
			},
		},
	}

	// POD spec for the stateful set.
	podSpec := corev1.PodSpec{
		Affinity: podAffinity,
		Containers: []corev1.Container{
			{
				Name:            ssName,
				Image:           dc.ZeroClusterSpec().Image(),
				ImagePullPolicy: dc.ZeroClusterSpec().PodImagePullPolicy(),
				Command: []string{
					"/bin/bash",
					"-c",
					zeroRunCmd,
				},
				Ports: []corev1.ContainerPort{
					{
						Name:          defaults.ZeroGRPCPortName,
						ContainerPort: defaults.ZeroGRPCPort,
						Protocol:      corev1.ProtocolTCP,
					},
					{
						Name:          defaults.ZeroHTTPPortName,
						ContainerPort: defaults.ZeroHTTPPort,
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
				VolumeMounts: podVolumeMounts,
				Resources:    dc.ZeroClusterSpec().ResourceRequirements(),
			},
		},
		RestartPolicy: corev1.RestartPolicyAlways,
	}

	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            ssName,
			Namespace:       ns,
			Labels:          zeroLabels,
			OwnerReferences: []metav1.OwnerReference{dc.AsOwnerReference()},
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicaCount,

			Selector: &metav1.LabelSelector{
				MatchLabels: zeroLabels,
			},
			PodManagementPolicy: appsv1.OrderedReadyPodManagement,
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{
					Partition: &partitionCount,
				}},

			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: zeroLabels,
				},
				Spec: podSpec,
			},
			ServiceName: headlessServiceName,
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: ssName,
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
						StorageClassName: &storageClassName,
						Resources: corev1.ResourceRequirements{
							Requests: dc.Spec.ZeroCluster.PersistentStorage.StorageRequest(),
						},
					},
				},
			},
		},
	}
}
