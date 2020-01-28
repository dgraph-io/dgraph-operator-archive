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

const (
	// AlphaMemberName is the component name of the alpha dgraph component.
	AlphaMemberName string = "dgraph-alpha"

	// ZeroMemberName is the component name of the Zero dgraph component.
	ZeroMemberName string = "dgraph-zero"

	// RatelMemberName is the component name of the Ratel dgraph component.
	RatelMemberName string = "dgraph-ratel"

	// AlphaMemberSuffix is the suffix to add to identifiers whenever required which
	// represents Dgraph alpha related components.
	AlphaMemberSuffix string = "alpha"

	// ZeroMemberSuffix is the suffix to add to identifiers whenever required which
	// represents Dgraph Zero related components.
	ZeroMemberSuffix string = "zero"

	// RatelMemberSuffix is the suffix to add to identifiers whenever required which
	// represents Dgraph ratel related components.
	RatelMemberSuffix string = "ratel"

	// ZeroGRPCPortName is the name of the port for Zero GRPC communication.
	ZeroGRPCPortName string = "zero-grpc"

	// ZeroGRPCPort is the port for dgraph zero GRPC communication.
	ZeroGRPCPort int32 = 5080

	// ZeroHTTPPortName is the name of the port for Zero HTTP communication.
	ZeroHTTPPortName string = "zero-http"

	// ZeroHTTPPort is the port for dgraph zero HTTP communication.
	ZeroHTTPPort int32 = 6080

	// HeadlessServiceSuffix is the suffix name to associate with dgraph headless services.
	HeadlessServiceSuffix string = "headless"

	// ZeroPersistentVolumeMountPath is the mount path for persistent volume that should be
	// attached to the zero container.
	ZeroPersistentVolumeMountPath string = "/dgraph"

	// AlphaPersistentVolumeMountPath is the mount path for persistent volume that should be
	// attached to the alpha container.
	AlphaPersistentVolumeMountPath string = "/dgraph"

	// AlphaGRPCPortName is the name of the port for Alpha GRPC communication.
	AlphaGRPCPortName string = "alpha-grpc"

	// AlphaGRPCPort is the port for dgraph Alpha GRPC communication.
	AlphaGRPCPort int32 = 9080

	// AlphaHTTPPortName is the name of the port for Alpha HTTP communication.
	AlphaHTTPPortName string = "alpha-http"

	// AlphaHTTPPort is the port for dgraph Alpha HTTP communication.
	AlphaHTTPPort int32 = 8080

	// MinLruMBValue is minimum value of LRUMb for alpha configuration.
	MinLruMBValue int32 = 512

	// LruMBValue is default value of LRU MB to use when a non valid value is provided
	// in the configuration.
	LruMBValue int32 = 2048

	// RatelPortName is the name of the port for Ratel UI.
	RatelPortName string = "ratel-grpc"

	// RatelPort is the port for dgraph Ratel UI.
	RatelPort int32 = 8000
)
