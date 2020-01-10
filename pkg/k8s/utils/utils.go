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

package utils

import (
	corev1 "k8s.io/api/core/v1"
)

// ResolveK8SServiceType returns kubernets service type corresponding to the type
// string provided as an argument.
func ResolveK8SServiceType(service string) corev1.ServiceType {
	switch service {
	case "NodePort":
		return corev1.ServiceTypeNodePort
	case "LoadBalancer":
		return corev1.ServiceTypeLoadBalancer
	default:
		return corev1.ServiceTypeClusterIP
	}
}
