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

package defaults

import (
	"time"
)

const (
	// CRDWaitPollInterval is the interval in which to regularly poll the K8s API server
	// for the availability of CRD.
	CRDWaitPollInterval time.Duration = 2 * time.Second

	// K8SAPIServerRequestTimeout is the default value of timeout for the request to
	// Kubernetes API server.
	K8SAPIServerRequestTimeout time.Duration = 20 * time.Second

	// InformerResyncDuration is the default resync duration of k8s shared informer factory.
	InformerResyncDuration time.Duration = 30 * time.Second

	// K8SDelimeter is the default delimeter for strings constructed for kubernetes context by
	// dgraph.
	K8SDelimeter string = "-"

	// StatefulSetPodAntiAffinityWeight is the weight of the weighted affinity term associated
	// with the dgraph stateful set component.
	StatefulSetPodAntiAffinityWeight int32 = 100

	// StatefulSetPodAntiAffinityKey is the topology key for the weighted affinity term associated
	// with the dgraph stateful set component.
	StatefulSetPodAntiAffinityKey string = "kubernetes.io/hostname"
)
