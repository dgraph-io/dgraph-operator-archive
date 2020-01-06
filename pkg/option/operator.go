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

package option

// operatorConfig is the configuration of dgraph operator.
//
// To connect to kubernetes API server the following preference order of
// parameters is used to create configuration:
// 1. KubeCfgPath
// 2. K8sAPIServerURL
// 3. From rest.InClusterConfig
type operatorConfig struct {
	// SkipCRDCreation skips the creation of CRDs required by dgraph operator.
	// If set to false, dgraph-operator will automatically create the required
	// Kubernetes custom resource definitions.
	SkipCRDCreation bool

	// K8sAPIServerURL is the kubernetes API server URL which can be explicitly
	// specified by the operator as a command line argument.
	K8sAPIServerURL string

	// KubeCfgPath is the path of kubernetes configuration file, which can be used
	// to create the configuration for rest client.
	KubeCfgPath string

	// Workers count is the number of workers to run for the controller.
	WorkersCount int

	// server represents configuration of the operator server.
	Server *operatorServer
}

// operatorServer is the server configuration for the operator.
type operatorServer struct {
	// Host to bind the server to.
	Host string

	// Port to bind the server to.
	Port int
}

// OperatorConfig is the variable which stores operator configuration.
var OperatorConfig = &operatorConfig{
	Server: &operatorServer{},
}
