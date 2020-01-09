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
	// OperatorHost is the default host for the operator server.
	OperatorHost string = "0.0.0.0"

	// OperatorPort is the default port that operator server listens to.
	OperatorPort int = 7777

	// LeaseLockName is the default value of lease lock we acquire when doing leader elections
	// among the operators.
	LeaseLockName string = "dgraph-io-controller-manager"

	// LeaderElectionLeaseDuration is the lease duration for leader election among the opearators.
	LeaderElectionLeaseDuration time.Duration = 15 * time.Second

	// LeaderElectionRenewDeadline is the renew deadline for the leader election among operators.
	LeaderElectionRenewDeadline time.Duration = 5 * time.Second

	// LeaderElectionRetryPeriod is the retry period of current operator for leader election
	// among operators.
	LeaderElectionRetryPeriod time.Duration = 3 * time.Second
)
