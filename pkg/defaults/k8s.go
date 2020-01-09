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
)
