package k8s

import (
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func int32Ptr(i int32) *int32 { return &i }

func operatorDeployment() *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dgraph-test-deployment",
		},
		// Run Deployment with multiple replicas
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "dgraph-operator-testing",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "dgraph-operator-testing",
					},
				},
				Spec: apiv1.PodSpec{
					ServiceAccountName: "dgraph-operator",
					Containers: []apiv1.Container{
						{
							Name:  "dgraph-operator",
							Image: Config.OperatorImage,
						},
					},
				},
			},
		},
	}
}

func createOperatorRBAC() error {
	rbacConfig := "./manifests/rbac.yaml"
	return Kubectl.Apply(rbacConfig)
}

func deleteOperatorRBAC() error {
	rbacConfig := "./manifests/rbac.yaml"
	return Kubectl.Delete(rbacConfig)
}
