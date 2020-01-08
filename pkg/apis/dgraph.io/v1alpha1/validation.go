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

package v1alpha1

import (
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

var (
	dgraphClusterCRV = &apiextv1.CustomResourceValidation{
		OpenAPIV3Schema: &apiextv1.JSONSchemaProps{
			Type: "object",
			Properties: map[string]apiextv1.JSONSchemaProps{
				"spec": dgraphClusterSchema,
			},
			Required: []string{"spec"},
		},
	}

	dgraphClusterSchema = apiextv1.JSONSchemaProps{
		Type: "object",
		Properties: map[string]apiextv1.JSONSchemaProps{
			"clusterID":       clusterIDSchema,
			"alpha":           alphaClusterSchema,
			"zero":            zeroClusterSchema,
			"ratel":           ratelClusterSchema,
			"baseImage":       dgraphComponentProperties["baseImage"],
			"version":         dgraphComponentProperties["version"],
			"imagePullPolicy": dgraphComponentProperties["imagePullPolicy"],
			"annotations":     dgraphComponentProperties["annotations"],
		},
		Required: []string{
			"clusterID",
			"alpha",
			"zero",
			"baseImage",
			"version",
		},
	}

	maxClusterIDLen int64 = 64
	clusterIDSchema       = apiextv1.JSONSchemaProps{
		Description: "Unique ID of the dgraph cluster deployment.",
		Type:        "string",
		MaxLength:   &maxClusterIDLen,
	}

	alphaClusterSchema = apiextv1.JSONSchemaProps{
		Description: "Configuration for dgraph alpha cluster.",
		Type:        "object",
		Required: []string{
			"replicas",
			"config",
		},
		Properties: map[string]apiextv1.JSONSchemaProps{
			"baseImage":       dgraphComponentProperties["baseImage"],
			"version":         dgraphComponentProperties["version"],
			"imagePullPolicy": dgraphComponentProperties["imagePullPolicy"],
			"annotations":     dgraphComponentProperties["annotations"],
			"requests":        dgraphComponentProperties["requests"],
			"limits":          dgraphComponentProperties["limits"],
			"replicas": {
				Description: "Number of replicas to run for alpha in the cluster.",
				Type:        "number",
			},
			"storageClassName": {
				Description: "Storage class to use for the component.",
				Type:        "string",
			},
			"config": {
				Description: "Config for dgraph alpha.",
				Type:        "object",
			},
		},
	}

	zeroClusterSchema = apiextv1.JSONSchemaProps{
		Description: "Configuration for dgraph zero cluster.",
		Type:        "object",
		Required: []string{
			"replicas",
			"config",
		},
		Properties: map[string]apiextv1.JSONSchemaProps{
			"baseImage":       dgraphComponentProperties["baseImage"],
			"version":         dgraphComponentProperties["version"],
			"imagePullPolicy": dgraphComponentProperties["imagePullPolicy"],
			"annotations":     dgraphComponentProperties["annotations"],
			"requests":        dgraphComponentProperties["requests"],
			"limits":          dgraphComponentProperties["limits"],

			"replicas": {
				Description: "Number of replicas to run for alpha in the cluster.",
				Type:        "number",
			},
			"storageClassName": {
				Description: "Storage class to use for the component.",
				Type:        "string",
			},
			"config": {
				Description: "Config for dgraph alpha.",
				Type:        "object",
			},
		},
	}

	ratelClusterSchema = apiextv1.JSONSchemaProps{
		Description: "Configuration for dgraph zero cluster.",
		Type:        "object",
		Required: []string{
			"replicas",
		},
		Properties: map[string]apiextv1.JSONSchemaProps{
			"baseImage":       dgraphComponentProperties["baseImage"],
			"version":         dgraphComponentProperties["version"],
			"imagePullPolicy": dgraphComponentProperties["imagePullPolicy"],
			"annotations":     dgraphComponentProperties["annotations"],
			"requests":        dgraphComponentProperties["requests"],
			"limits":          dgraphComponentProperties["limits"],

			"replicas": {
				Description: "Number of replicas to run for ratel in the cluster.",
				Type:        "number",
			},
		},
	}

	dgraphComponentProperties = map[string]apiextv1.JSONSchemaProps{
		"baseImage": {
			Description: "Base image(without tag) to use for dgraph component cluster.",
			Type:        "string",
		},
		"version": {
			Description: "Version of the dgraph component to use in the clsuter.",
			Type:        "string",
		},
		"imagePullPolicy": {
			Description: "Image pull policy for the containers to run.",
			Type:        "string",
		},
		"annotations": {
			Description: "Annotations to apply on the kubernetes container object.",
			Type:        "object",
		},
		"limits": {
			Description: "Limits describes the maximum amount of compute resources " +
				"allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/.",
			Type: "object",
		},
		"requests": {
			Description: "Requests describes the minimum amount of compute resources" +
				"required. If Requests is omitted for a container, it defaults" +
				"to Limits if that is explicitly specified, otherwise to an implementation-defined" +
				"value. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/",
			Type: "object",
		},
	}
)
