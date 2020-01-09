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

	dc "github.com/dgraph-io/dgraph-operator/pkg/controller/dgraphcluster"
	"github.com/dgraph-io/dgraph-operator/pkg/defaults"
	"github.com/dgraph-io/dgraph-operator/pkg/k8s"

	"github.com/golang/glog"
	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
)

// Controller is the standard interface which each controller within dgraph operator
// must implement.
type Controller interface {
	Run(context.Context)
}

var (
	registeredControllers = []Controller{
		dc.NewController(),
	}
)

// RunOperatorControllers runs all the required controllers for dgraph operator.
// Each controller watches for the resources it reconciles and take necessery action
// to reach to the desired state of the resource.
//
// Here we also implement the logic of leader election for dgraph operator using
// built-in leader election capbility in kubernetes.
// See: https://github.com/kubernetes/client-go/blob/master/examples/leader-election/main.go
func RunOperatorControllers() error {
	client, err := k8s.Client()
	if err != nil {
		return err
	}

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
	//    Lease.coordination.k8s.io "dgraph-io-controller-manager" is invalid: spec.leaseDurationSeconds:
	//    Invalid value: 0: must be greater than 0
	//
	// TODO: Use LeaseLock here.
	leaseLock := &resourcelock.EndpointsLock{
		EndpointsMeta: metav1.ObjectMeta{
			Name:      defaults.LeaseLockName,
			Namespace: ns,
		},
		Client: client.CoreV1(),
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
			OnStartedLeading: onStart,
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
func onStart(ctx context.Context) {
	glog.Info("started leading.")

	// Start running managed controllers here
	for _, ctrl := range registeredControllers {
		go ctrl.Run(ctx)
	}
	select {}
}
