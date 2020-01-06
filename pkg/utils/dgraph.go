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
	"fmt"

	"github.com/dgraph-io/dgraph-operator/pkg/defaults"
)

// DgraphAlphaMemberName is the name of alpha member associated with cluster provided.
// The format is <clusterID>-<clusterName>-alpha
func DgraphAlphaMemberName(clusterID, clusterName string) string {
	return fmt.Sprintf("%s%s%s%s%s",
		clusterID, defaults.K8SDelimeter, clusterName, defaults.K8SDelimeter, defaults.AlphaMemberSuffix)
}

// DgraphZeroMemberName is the name of Zero member associated with cluster provided.
// The format is <clusterID>-<clusterName>-zero
func DgraphZeroMemberName(clusterID, clusterName string) string {
	return fmt.Sprintf("%s%s%s%s%s",
		clusterID, defaults.K8SDelimeter, clusterName, defaults.K8SDelimeter, defaults.ZeroMemberSuffix)
}

// DgraphRatelMemberName is the name of ratel member associated with cluster provided.
// The format is <clusterID>-<clusterName>-ratel
func DgraphRatelMemberName(clusterID, clusterName string) string {
	return fmt.Sprintf("%s%s%s%s%s",
		clusterID, defaults.K8SDelimeter, clusterName, defaults.K8SDelimeter, defaults.RatelMemberSuffix)
}
