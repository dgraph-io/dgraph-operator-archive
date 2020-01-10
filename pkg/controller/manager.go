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

package controller

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/dgraph-io/dgraph-operator/pkg/client/clientset/versioned"
	informers "github.com/dgraph-io/dgraph-operator/pkg/client/informers/externalversions"
	dc "github.com/dgraph-io/dgraph-operator/pkg/controller/dgraphcluster"
	"github.com/dgraph-io/dgraph-operator/pkg/defaults"
	"github.com/dgraph-io/dgraph-operator/pkg/k8s"

	"github.com/golang/glog"
	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
)

// Manager manages all the controllers for the operator.
// This is the top level manager for all the underlying controllers and managers.
// All controllers which operator manages should be in the registered controller list
// of Manager.
type Manager struct {
	k8sClient    kubernetes.Interface
	dgraphClient versioned.Interface

	registeredControllers []Controller
}

// Controller is the standard interface which each controller within dgraph operator
// must implement.
// List of configured controllers managed by dgraph operator as of now are:
// * DgraphClusterController
type Controller interface {
	// Run starts running the controller watching for required kubernetes resources
	// and associating required handler with resource events.
	Run(context.Context)
}

// MustNewControllerManager returns a Manager instance if it is able
// to do so, else it exists the program.
func MustNewControllerManager() *Manager {
	k8sClient, err := k8s.Client()
	if err != nil {
		glog.Fatalf("error while building kubernetes client.")
	}

	dgraphClient, err := k8s.DgraphClient()
	if err != nil {
		glog.Fatalf("error while building dgraph k8s client")
	}

	// We register controller during manager run. This is to make sure that
	// Different controller can also share the same informer factory
	// from kubernetes.
	registeredControllers := make([]Controller, 0)

	return &Manager{
		k8sClient,
		dgraphClient,

		registeredControllers,
	}
}

// RunOperatorControllers runs all the required controllers for dgraph operator.
// Each controller watches for the resources it reconciles and take necessery action
// to reach to the desired state of the resource.
//
// Here we also implement the logic of leader election for dgraph operator using
// built-in leader election capbility in kubernetes.
// See: https://github.com/kubernetes/client-go/blob/master/examples/leader-election/main.go
func (cm *Manager) RunOperatorControllers() error {
	// Get hostname for identity name of the lease lock holder.
	// We identify the leader of the operator cluster using hostname.
	// If there is an error while getting hostname we use a randomly generated
	// UUID as the leader name.
	hostID, err := os.Hostname()
	if err != nil {
		glog.Errorf("failed to get hostname: %s", err)
		hostID = uuid.New().String()
	}
	glog.Infof("using host ID: %s", hostID)

	ns := os.Getenv("NAMESPACE")
	// If due to any reason the NAMESPACE is not set we assume it to be
	// in default namespace.
	if ns == "" {
		ns = "default"
	}
	// Use a Go context so we can tell the leaderelection code when we
	// want to step down
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Listen for interrupts or the Linux SIGTERM signal and cancel
	// our context, which the leader election code will observe and
	// step down.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		glog.Info("received termination, signaling shutdown.")
		cancel()
	}()

	// We should use the Lease lock type since edits to Leases are less common
	// and fewer objects in the cluster watch "all Leases".
	// There is an issue with leaselocks for now which gives the following error:
	// E0102 16:45:18.120239   24632 leaderelection.go:307] Failed to release lock:
	//    Lease.coordination.k8s.io "dgraph-io-controller-manager" is invalid:
	// 	  spec.leaseDurationSeconds:
	//    Invalid value: 0: must be greater than 0
	//
	// TODO: Use LeaseLock here.
	leaseLock := &resourcelock.EndpointsLock{
		EndpointsMeta: metav1.ObjectMeta{
			Name:      defaults.LeaseLockName,
			Namespace: ns,
		},
		Client: cm.k8sClient.CoreV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			// Identity name of theholder
			Identity:      hostID,
			EventRecorder: &record.FakeRecorder{},
		},
	}

	// Start the leader election for running dgraph operators
	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:            leaseLock,
		ReleaseOnCancel: true,

		LeaseDuration: defaults.LeaderElectionLeaseDuration,
		RenewDeadline: defaults.LeaderElectionRenewDeadline,
		RetryPeriod:   defaults.LeaderElectionRetryPeriod,

		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: cm.onStart,
			OnStoppedLeading: func() {
				glog.Infof("leader election lost: %s", hostID)
			},
			OnNewLeader: func(identity string) {
				if identity == hostID {
					// We are the new leader.
					return
				}
				glog.Infof("new leader elected: %s", identity)
			},
		},
	})

	return nil
}

// onStart is the main function which is executed when our operator starts leading
// the cluster of operators in the kubernetes cluster.
// onStart also populates the registered controller that the operator manages.
func (cm *Manager) onStart(ctx context.Context) {
	glog.Info("started leading.")
	glog.Info("registering controllers for operator")
	defer func() {
		cm.registeredControllers = make([]Controller, 0)
		glog.Info("exitting onStart method...")
	}()

	dgraphClusterInformer := informers.NewSharedInformerFactory(
		cm.dgraphClient,
		defaults.InformerResyncDuration)
	k8sInformerFactory := k8sinformers.NewSharedInformerFactory(
		cm.k8sClient,
		defaults.InformerResyncDuration)

	// Add dgraph controller to registered controller list of the controller manager.
	cm.registeredControllers = append(cm.registeredControllers, dc.NewController(
		cm.k8sClient,
		cm.dgraphClient,
		dgraphClusterInformer.Dgraph().V1alpha1().DgraphClusters(),
		k8sInformerFactory,
	))

	// notice that there is no need to run Start methods in a separate goroutine.
	// (i.e. go informerFactory.Start(stopCh) Start method is non-blocking and
	// runs all registered informers in a dedicated goroutine.
	dgraphClusterInformer.Start(ctx.Done())
	k8sInformerFactory.Start(ctx.Done())

	// Iterate through all the registered controllers and run them in a
	// separate goroutine.
	for _, ctrl := range cm.registeredControllers {
		// run the required controller.
		go ctrl.Run(ctx)
	}

	glog.Info("all controllers configured and are now running.")
	<-ctx.Done()
}
