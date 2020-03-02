package k8s

import "time"

const (
	// defaultKubeconfigPath is the default path to use for kubeconfig file
	defaultKubeconfigPath string = "/home/fristonio/.kube/config"

	// defaultOperatorImage is the default operator image to use for running the operator.
	defaultOperatorImage string = "localhost:5000/dgraph-io/dgraph-operator"

	// defaultRequestTimeout is the timeout duration for kubernetes api server requests.
	defaultRequestTimeout time.Duration = 30 * time.Second

	// pollInterval is the default value of polling interval when watching resources.
	pollInterval time.Duration = 15 * time.Second

	// kubectlActionTimeout is the timeout duration for a kubectl action command
	kubectlActionTimeout time.Duration = 30 * time.Second

	// dgraphClusterCRD is the name of dgraph cluster CRD.
	dgraphClusterCRD string = "dgraphclusters.dgraph.io"
)

// TestConfig is the configuration structure for running dgraph operator tests
// on kubernetes.
type TestConfig struct {
	// OperatorImage is the image to use for the dgraph-operator, this can be overridden.
	OperatorImage string

	// Kubeconfig is the path of the kubeconfig file to use for interacting with the
	// kubernetes cluster.
	Kubeconfig string

	// Context to use with kubectl.
	Context string
}

// Config is the config data for running the test, populated using command line
// flags.
var Config = &TestConfig{}
