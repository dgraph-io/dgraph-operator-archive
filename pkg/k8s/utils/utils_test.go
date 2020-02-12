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
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

func TestResolveK8SServiceType(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(ResolveK8SServiceType("NodePort")).To(Equal(corev1.ServiceTypeNodePort))
	g.Expect(ResolveK8SServiceType("LoadBalancer")).To(Equal(corev1.ServiceTypeLoadBalancer))
	g.Expect(ResolveK8SServiceType("ClusterIP")).To(Equal(corev1.ServiceTypeClusterIP))
}
