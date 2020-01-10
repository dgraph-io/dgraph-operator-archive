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

package dgraphcluster

import (
	"context"
	"fmt"
	"time"

	dgraphio "github.com/dgraph-io/dgraph-operator/pkg/apis/dgraph.io/v1alpha1"
	"github.com/dgraph-io/dgraph-operator/pkg/client/clientset/versioned"
	dgraphscheme "github.com/dgraph-io/dgraph-operator/pkg/client/clientset/versioned/scheme"
	// nolint
	dgraphinformer "github.com/dgraph-io/dgraph-operator/pkg/client/informers/externalversions/dgraph.io/v1alpha1"
	listers "github.com/dgraph-io/dgraph-operator/pkg/client/listers/dgraph.io/v1alpha1"
	"github.com/dgraph-io/dgraph-operator/pkg/defaults"
	"github.com/dgraph-io/dgraph-operator/pkg/k8s"
	"github.com/dgraph-io/dgraph-operator/pkg/manager"
	"github.com/dgraph-io/dgraph-operator/pkg/option"

	"github.com/golang/glog"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	k8sinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

// Controller is the controller to manage the DgraphCluster custom
// resource created in the Kubernetes cluster.
//
// Controller type uses bits and pieces from here:
// https://github.com/kubernetes/sample-controller/
type Controller struct {
	// k8sClient is the client interface to connect to the kube API server.
	k8sClient kubernetes.Interface

	// dgraphClient is the client interface to interacting with dgraph related
	// custom resources.
	dgraphClient versioned.Interface

	dgraphClusterLister listers.DgraphClusterLister
	dgraphClusterSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface

	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder

	// Each controller contains a list of Managers with it. These managers are responsible
	// for managing(in our case syncing) the managed resources in accordance with the
	// latest version of custom resource they are managing.
	// These manager shares the same kubernetes listers to interact with the API server.
	// These managers are run sequentially and thus must be present in the order required
	// for underlying resources.
	// For example in case of DgraphCluster we have three resources managers:
	// * AlphaManager
	// * ZeroManager
	// * RatelManager
	//
	// For a proper dgraph cluster provisioning we assume to ensure the following
	// order in their individual syncs:
	// Zero -> Alpha -> Ratel
	// and they should be present in this particular order in the managers list.
	managers []manager.Manager
}

// NewController returns a new DgraphCluster controller.
func NewController(k8sClient kubernetes.Interface,
	dgraphClient versioned.Interface,
	dgraphClusterInformer dgraphinformer.DgraphClusterInformer,
	k8sInformerFactory k8sinformers.SharedInformerFactory) *Controller {

	utilruntime.Must(dgraphscheme.AddToScheme(scheme.Scheme))
	glog.Info("Creating event broadcaster")

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(
		&typedcorev1.
			EventSinkImpl{Interface: k8sClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(
		scheme.Scheme,
		v1.EventSource{Component: defaults.DgraphOperatorName})

	workqueue := workqueue.NewNamedRateLimitingQueue(
		workqueue.DefaultControllerRateLimiter(),
		"DgraphClusters")

	ctrl := &Controller{
		k8sClient:    k8sClient,
		dgraphClient: dgraphClient,

		recorder:  recorder,
		workqueue: workqueue,
	}

	ctrl.dgraphClusterLister = dgraphClusterInformer.Lister()
	ctrl.dgraphClusterSynced = dgraphClusterInformer.Informer().HasSynced

	// event handlers for DgraphCluster custom kubernetes resource.
	dgraphClusterInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			glog.Info("dgraph-cluster-controller: add on DgraphCluster CRD invoked.")
			ctrl.enqueueObj(obj)
		},
		UpdateFunc: func(old, cur interface{}) {
			glog.Info("dgraph-cluster-controller: update on DgraphClsuter CRD invoked.")
			ctrl.enqueueObj(cur)
		},
		DeleteFunc: func(obj interface{}) {
			glog.Info("dgraph-cluster-controller: delete on DgraphCluster CRD invoked.")
			ctrl.enqueueObj(obj)
		},
	})

	// Listers for required kubernetes resources.
	podsLister := k8sInformerFactory.Core().V1().Pods().Lister()
	svcLister := k8sInformerFactory.Core().V1().Services().Lister()
	statefulSetLister := k8sInformerFactory.Apps().V1().StatefulSets().Lister()
	deploymentLister := k8sInformerFactory.Apps().V1().Deployments().Lister()

	// setup managers for DgraphCluster resources.
	// These managers must be synced in this particular order only.
	// Zero -> Alpha -> Ratel
	managers := make([]manager.Manager, 0)
	managers = append(managers, manager.NewZeroManager(
		k8sClient,
		podsLister,
		svcLister,
		statefulSetLister,
	))
	managers = append(managers, manager.NewAlphaManager(
		k8sClient,
		podsLister,
		svcLister,
		statefulSetLister,
	))
	managers = append(managers, manager.NewRatelManager(
		k8sClient,
		podsLister,
		svcLister,
		deploymentLister,
	))

	ctrl.managers = managers

	return ctrl
}

// Run runs the actual underlying DgraphCluster controller.
func (dc *Controller) Run(ctx context.Context) {
	glog.Info("dgraph-cluster-controller: starting to run DgraphCluster controller")

	// Kubernetes specific controller teardown logic.
	defer utilruntime.HandleCrash()
	defer dc.workqueue.ShutDown()

	// Wait for CRD to be ready, skip if any error occurs.
	if err := k8s.WaitForCRD(dgraphio.DgraphClusterCRDName); err != nil {
		glog.Warningf("dgraph-cluster-controller: error while waiting for CRD "+
			"to be ready: %s\nignoring failure", err)
	}

	glog.Info("dgraph-cluster-controller: waiting for informer cache to sync")
	if ok := cache.WaitForCacheSync(ctx.Done(), dc.dgraphClusterSynced); !ok {
		glog.Fatalf("dgraph-cluster-controller: error while syncing informer cache, exitting")
	}
	glog.Info("dgraph-cluster-controller: informer cache synced.")

	// Run WorkersCount number of workers to process the work from the queue.
	for i := 0; i < option.OperatorConfig.WorkersCount; i++ {
		go wait.Until(dc.runWorker, time.Second, ctx.Done())
	}

	glog.Info("dgraph-cluster-controller: started workers")
	<-ctx.Done()
	glog.Info("dgraph-cluster-controller: shutting down workers")
}

func (dc *Controller) runWorker() {
	for dc.processNextWorkItem() {
		// long-running function that will continually call the
		// processNextWorkItem function in order to read and process
		// a message on the workqueue.
	}
}

// process a work item from the workqueue.
func (dc *Controller) processNextWorkItem() bool {
	obj, shutdown := dc.workqueue.Get()

	// Got shutdown when getting object from workqueue.
	// We exit from processing the work items.
	if shutdown {
		return false
	}

	// We call Done here so the workqueue knows we have finished
	// processing this item. We also must remember to call Forget if we
	// do not want this work item being re-queued. For example, we do
	// not call Forget if a transient error occurs, instead the item is
	// put back on the workqueue and attempted again after a back-off
	// period.
	defer dc.workqueue.Done(obj)
	var objKey string
	var ok bool

	// We expect strings to come off the workqueue. These are of the
	// form namespace/name. We do this as the delayed nature of the
	// workqueue means the items in the informer cache may actually be
	// more up to date that when the item was initially put onto the
	// workqueue.
	if objKey, ok = obj.(string); !ok {
		// As the item in the workqueue is actually invalid, we call
		// Forget here else we'd go into a loop of attempting to
		// process a work item that is invalid.
		dc.workqueue.Forget(obj)
		utilruntime.HandleError(fmt.Errorf("dgraph-cluster-controller: expected string in "+
			"workqueue but got %#v", obj))
		return true
	}

	// Sync the object
	if err := dc.sync(objKey); err != nil {
		// Put the item back on the workqueue to handle any transient errors.
		// This item we be retried at later point of time.
		dc.workqueue.AddRateLimited(objKey)
		err = fmt.Errorf("dgraph-cluster-controller: error syncing '%s': %s, requeuing",
			objKey,
			err.Error())
	} else {
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		dc.workqueue.Forget(obj)
	}
	glog.Infof("dgraph-cluster-controller: successfully synced '%s'", objKey)

	return true
}

// Syncs the DgraphCluster resource represented by `key`
func (dc *Controller) sync(key string) error {
	startTime := time.Now()
	defer glog.Infof("dgraph-cluster-controller: DgraphCluster sync done %q (%v)",
		key,
		time.Since(startTime))

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	// Get the latest DgraphCluster CRD object from the API server we are syncing here.
	cluster, err := dc.dgraphClusterLister.DgraphClusters(namespace).Get(name)
	if kerrors.IsNotFound(err) {
		// If a dgraph cluster has already been deleted then we don't do anything
		// this is because all the resources we create using dgraph controllers
		// have owner reference attached to this particular type.
		// So once an object is deleted, all the additional resources associated
		// with it are also deleted.
		glog.Infof("dgraph-cluster-controller: DgraphCluster(%q) has already been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	// Update the dgraph cluster based on the latest object we got from the
	// kubernetes API.
	// Each controller similar to DgraphCluster one must implement an update function which
	// updates the underlying resources based on the latest configuration
	// we got from the kubernetes API server.
	err = dc.UpdateDgraphCluster(cluster.DeepCopy())
	if err != nil {
		glog.Errorf("dgraph-cluster-controller: error while updating dgraph cluster "+
			"with provided specification: %s", err)
	}

	return err
}

// enqueueObj enqueues the object to the work queue.
func (dc *Controller) enqueueObj(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("dgraph-cluster-controller: cound't get "+
			"key for object %+v: %v", obj, err))
		return
	}
	glog.Infof("dgraph-cluster-controller: enqueuing %q in workqueue", key)
	dc.workqueue.Add(key)
}
