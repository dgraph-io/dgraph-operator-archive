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

// DefaultAlphaLabels returns a map representing labels associated with dgraph Alpha component.
func DefaultAlphaLabels(instanceName string) map[string]string {
	AlphaLabels := labels.NewLabelSet().
		Instance(instanceName).
		Component(defaults.AlphaMemberName).
		ManagedBy(defaults.DgraphOperatorName)

	return AlphaLabels
}

// NewAlphaService constructs a K8s service object for dgraph Alpha from the provided DgraphCluster
// configuration.
func NewAlphaService(dc *v1alpha1.DgraphCluster) *corev1.Service {
	ns := dc.GetNamespace()
	name := dc.GetName()
	clusterID := dc.Spec.GetClusterID()

	serviceName := utils.DgraphAlphaMemberName(clusterID, name)
	alphaLabels := DefaultAlphaLabels(serviceName)

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            serviceName,
			Namespace:       ns,
			Labels:          alphaLabels,
			OwnerReferences: []metav1.OwnerReference{dc.AsOwnerReference()},
		},
		Spec: corev1.ServiceSpec{
			Type: dc.Spec.AlphaServiceType(),
			Ports: []corev1.ServicePort{
				{
					Name:       defaults.AlphaGRPCPortName,
					Port:       defaults.AlphaGRPCPort,
					TargetPort: intstr.FromInt(int(defaults.AlphaGRPCPort)),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       defaults.AlphaHTTPPortName,
					Port:       defaults.AlphaHTTPPort,
					TargetPort: intstr.FromInt(int(defaults.AlphaHTTPPort)),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: alphaLabels,
		},
	}
}

// NewAlphaHeadlessService constructs a K8s headless service object for dgraph Alpha
// from the provided DgraphCluster configuration.
func NewAlphaHeadlessService(dc *v1alpha1.DgraphCluster) *corev1.Service {
	svc := NewAlphaService(dc)

	// Service name for headless service is of the format
	// <clusterID>-<clusterName>-alpha-headless
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

// NewAlphaStatefulSet constructs a K8s stateful set object for dgraph Alpha from the
// provided DgraphCluster configuration.
func NewAlphaStatefulSet(dc *v1alpha1.DgraphCluster) *appsv1.StatefulSet {
	ns := dc.GetNamespace()
	name := dc.GetName()
	clusterID := dc.Spec.GetClusterID()

	ssName := utils.DgraphAlphaMemberName(clusterID, name)
	zeroMemberName := utils.DgraphZeroMemberName(clusterID, name)
	headlessServiceName := fmt.Sprintf("%s%s%s",
		ssName,
		defaults.K8SDelimeter,
		defaults.HeadlessServiceSuffix)
	storageClassName := dc.Spec.AlphaCluster.PersistentStorage.StorageClassName

	lruMB := dc.Spec.AlphaCluster.LruMB()
	if lruMB < defaults.MinLruMBValue {
		lruMB = defaults.LruMBValue
	}
	alphaLabels := DefaultAlphaLabels(ssName)

	replicaCount := dc.Spec.AlphaCluster.Replicas
	partitionCount := dc.Spec.AlphaCluster.Replicas
	// nolint
	AlphaRunCmd := fmt.Sprintf(`set -ex
dgraph alpha --my=$(hostname -f):7080 --lru_mb %d --zero %s-0.%s-headless.${POD_NAMESPACE}.svc.cluster.local:5080
`, lruMB, zeroMemberName, zeroMemberName)

	podVolumeMounts := []corev1.VolumeMount{
		{
			Name:      ssName,
			MountPath: defaults.AlphaPersistentVolumeMountPath,
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
							MatchLabels: labels.NewLabelSet().Component(defaults.AlphaMemberName),
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
				Image:           dc.AlphaClusterSpec().Image(),
				ImagePullPolicy: dc.AlphaClusterSpec().PodImagePullPolicy(),
				Command: []string{
					"/bin/bash",
					"-c",
					AlphaRunCmd,
				},
				Ports: []corev1.ContainerPort{
					{
						Name:          defaults.AlphaGRPCPortName,
						ContainerPort: defaults.AlphaGRPCPort,
						Protocol:      corev1.ProtocolTCP,
					},
					{
						Name:          defaults.AlphaHTTPPortName,
						ContainerPort: defaults.AlphaHTTPPort,
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
				Resources:    dc.AlphaClusterSpec().ResourceRequirements(),
			},
		},
		RestartPolicy: corev1.RestartPolicyAlways,
	}

	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            ssName,
			Namespace:       ns,
			Labels:          alphaLabels,
			OwnerReferences: []metav1.OwnerReference{dc.AsOwnerReference()},
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicaCount,

			Selector: &metav1.LabelSelector{
				MatchLabels: alphaLabels,
			},
			PodManagementPolicy: appsv1.OrderedReadyPodManagement,
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{
					Partition: &partitionCount,
				}},

			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: alphaLabels,
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
							Requests: dc.Spec.AlphaCluster.PersistentStorage.StorageRequest(),
						},
					},
				},
			},
		},
	}
}
