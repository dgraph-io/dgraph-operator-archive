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

package k8s

import (
	"net"
	"os"
	"strings"

	"github.com/dgraph-io/dgraph-operator/pkg/client/clientset/versioned"
	"github.com/dgraph-io/dgraph-operator/pkg/option"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// These client wrappers are necessery to add function specific to dgraph operator
// like annotating each resource with DgraphCluster ID etc.

// ClientK8s is a wrapper around kubernetes.Interface.
type ClientK8s struct {
	// kubernetes.Interface is the object through which interactions with
	// Kubernetes are performed.
	kubernetes.Interface
}

// DgraphClientK8s is a wrapper around clientset.Interface for dgraph
// operator kubernetes client.
type DgraphClientK8s struct {
	versioned.Interface
}

// Client creates a new k8s client
func Client() (*ClientK8s, error) {
	cfg, err := CreateConfig(option.OperatorConfig.K8sAPIServerURL,
		option.OperatorConfig.KubeCfgPath)
	if err != nil {
		return nil, err
	}
	c, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &ClientK8s{Interface: c}, nil
}

// DgraphClient creates a new k8s api extensions client
func DgraphClient() (*DgraphClientK8s, error) {
	cfg, err := CreateConfig(option.OperatorConfig.K8sAPIServerURL,
		option.OperatorConfig.KubeCfgPath)
	if err != nil {
		return nil, err
	}
	c, err := versioned.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &DgraphClientK8s{Interface: c}, nil
}

// APIExtClient creates a new k8s api extensions client
func APIExtClient() (apiextensionsclient.Interface, error) {
	cfg, err := CreateConfig(option.OperatorConfig.K8sAPIServerURL,
		option.OperatorConfig.KubeCfgPath)
	if err != nil {
		return nil, err
	}
	c, err := apiextensionsclient.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// CreateConfig creates a rest.Config for connecting to k8s api-server.
//
// The precedence of the configuration selection is the following:
// 1. kubeCfgPath
// 2. apiServerURL (https if specified)
// 3. rest.InClusterConfig().
func CreateConfig(apiServerURL, kubeCfgPath string) (*rest.Config, error) {
	var (
		config *rest.Config
		err    error
	)

	switch {
	// If the apiServerURL and the kubeCfgPath are empty then we can try getting
	// the rest.Config from the InClusterConfig
	case apiServerURL == "" && kubeCfgPath == "":
		if config, err = inClusterConfig(); err != nil {
			return nil, err
		}
	case kubeCfgPath != "":
		if config, err = clientcmd.BuildConfigFromFlags("", kubeCfgPath); err != nil {
			return nil, err
		}
	case strings.HasPrefix(apiServerURL, "https://"):
		if config, err = rest.InClusterConfig(); err != nil {
			return nil, err
		}
		config.Host = apiServerURL
	default:
		config = &rest.Config{Host: apiServerURL}
	}

	return config, nil
}

// InClusterConfig loads the environment into a rest config.
func inClusterConfig() (*rest.Config, error) {
	if len(os.Getenv("KUBERNETES_SERVICE_HOST")) == 0 {
		addrs, err := net.LookupHost("kubernetes.default.svc")
		if err != nil {
			return nil, err
		}
		os.Setenv("KUBERNETES_SERVICE_HOST", addrs[0])
	}
	if len(os.Getenv("KUBERNETES_SERVICE_PORT")) == 0 {
		os.Setenv("KUBERNETES_SERVICE_PORT", "443")
	}
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
