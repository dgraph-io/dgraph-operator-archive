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
	AlphaMemberName string = "alpha"

	// ZeroMemberName is the component name of the Zero dgraph component.
	ZeroMemberName string = "zero"

	// RatelMemberName is the component name of the Ratel dgraph component.
	RatelMemberName string = "ratel"

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
)
