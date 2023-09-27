/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/dgraph-io/dgraph-operator/api/v1alpha1"
)

type DgraphNodeType string

const (
	Alpha DgraphNodeType = "alpha"
	Zero  DgraphNodeType = "zero"

	AlphaLabel = "dgraph-alpha"
	ZeroLabel  = "dgraph-zero"

	AlphaStatefulSet = "alphastatefulset"
	ZeroStatefulSet  = "zerostatefulset"

	AlphaService  = "dgraph-alpha-service"
	PublicService = "dgraph-service"
	ZeroService   = "dgraph-zero-service"
)

type clusterResources struct {
	appLabel             string
	svcName              string
	podTemplate          corev1.PodTemplateSpec
	volumeClaimTemplates []corev1.PersistentVolumeClaim
}

// DgraphClusterReconciler reconciles a DgraphCluster object
type DgraphClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=dgraph.io,resources=dgraphclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dgraph.io,resources=dgraphclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dgraph.io,resources=dgraphclusters/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=statefulsets/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DgraphCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.0/pkg/reconcile
func (r *DgraphClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues("req", req)
	log.Info("reconciling DgraphCluster")

	// Fetch the DgraphCluster instance's desired state (spec)
	var dgc v1alpha1.DgraphCluster
	if err := r.Get(ctx, req.NamespacedName, &dgc); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue.
			log.Info("DgraphCluster not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "error reading DgraphCluster resource")
		return ctrl.Result{}, err
	}

	requeue, err := r.syncStatefulSet(ctx, dgc, Zero)
	if err != nil || requeue {
		log.Error(err)
		return ctrl.Result{Requeue: requeue}, client.IgnoreNotFound(err)
	}
	log.Info("Zero done. Alpha next")
	requeue, err = r.syncStatefulSet(ctx, dgc, Alpha)
	if err != nil || requeue {
		return ctrl.Result{Requeue: requeue}, client.IgnoreNotFound(err)
	}

	log.Info("reconciled DgraphCluster")
	return ctrl.Result{}, nil
}

func (r *DgraphClusterReconciler) syncStatefulSet(ctx context.Context, dgc v1alpha1.DgraphCluster, nodeType DgraphNodeType) (bool, error) {
	log := log.FromContext(ctx).WithValues()
	name := map[DgraphNodeType]string{
		"alpha": AlphaStatefulSet,
		"zero":  ZeroStatefulSet,
	}[nodeType]

	// Check if sts already exists, if not create a new one
	found := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{Namespace: dgc.Namespace, Name: name}, found)
	if err != nil && errors.IsNotFound(err) {
		sts := buildStatefulSet(dgc, nodeType, name)
		// set DgraphCluster instance as the owner and controller
		if err := controllerutil.SetControllerReference(&dgc, sts, r.Scheme); err != nil {
			return false, err
		}

		log.Info("creating new StatefulSet", "namespace", sts.Namespace, "name", name)
		err = r.Create(ctx, sts)
		if err != nil {
			log.Error(err, "failed to create new StatefulSet", "namespace", sts.Namespace, "name", name)
			return false, err
		}

		return true, nil
	} else if err != nil {
		log.Error(err, "failed to get StatefulSet")
		return false, err
	}

	// Ensure the sts size is the same as the spec
	size := dgc.Spec.Size
	if found.Spec.Replicas != size {
		found.Spec.Replicas = size
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "failed to update StatefulSet", "namespace", found.Namespace, "name", found.Name)
			return false, err
		}
		// Spec updated - return and requeue
		return true, nil
	}

	return false, nil
}

func buildStatefulSet(dgc v1alpha1.DgraphCluster, nodeType DgraphNodeType, name string) *appsv1.StatefulSet {
	cr := map[DgraphNodeType]clusterResources{
		"alpha": clusterResources{
			appLabel:             AlphaLabel,
			svcName:              AlphaService,
			podTemplate:          dgc.Spec.AlphaPod.PodTemplate,
			volumeClaimTemplates: dgc.Spec.AlphaPod.VolumeClaimTemplates,
		},
		"zero": clusterResources{
			appLabel:             ZeroLabel,
			svcName:              ZeroService,
			podTemplate:          dgc.Spec.ZeroPod.PodTemplate,
			volumeClaimTemplates: dgc.Spec.ZeroPod.VolumeClaimTemplates,
		},
	}[nodeType]

	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: dgc.Namespace,
			Name:      name,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:            dgc.Spec.Size,
			PodManagementPolicy: appsv1.OrderedReadyPodManagement,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": cr.appLabel,
				},
			},
			ServiceName:          cr.svcName,
			Template:             cr.podTemplate,
			VolumeClaimTemplates: cr.volumeClaimTemplates,
		},
	}
	return sts
}

// SetupWithManager sets up the controller with the Manager.
func (r *DgraphClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.DgraphCluster{}).
		Owns(&appsv1.StatefulSet{}).
		Complete(r)
}
