package k8s

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// KubectlCmd is the command to use for kubectl
	KubectlCmd = "kubectl"

	kubectlApplyCmd  = "apply"
	kubectlDeleteCmd = "delete"

	// KubectlContextEnv is the environment variable name which is used to
	// get the context to use when executing kubectl commands.
	KubectlContextEnv = "KUBECTL_CONTEXT"
)

type kubectl struct{}

// CmdRes is the response of an executed command, it might contain more structured
// information related to the executed command.
// This is a custom type we use to represent the result of an executed command
// This response type is associated with a few helper method which makes
// it easy to perform some action based on the result of the command executed.
type CmdRes struct {
	Cmd    string
	stdout bytes.Buffer
	stderr bytes.Buffer

	exitcode int
	err      error

	Duration time.Duration
}

// ExitCode returns the exit code of the command executed.
func (r *CmdRes) ExitCode() int {
	return r.exitcode
}

// GetStdout returns the string representation of the standard output.
func (r *CmdRes) GetStdout() string {
	return r.stdout.String()
}

// GetStderr returns the string representation of the standard error output.
func (r *CmdRes) GetStderr() string {
	return r.stderr.String()
}

// Error returns the error if any which occured during the execution of the
// provided command.
func (r *CmdRes) Error() error {
	return r.err
}

// Kubectl is the variable which will be configured with the required
// configuration which can be used to interact with our kubernetes cluster
// for testing purposes.)
var Kubectl = &kubectl{}

// Cmd executes a kubectl command on the local node.
func (k *kubectl) RunCmd(ctx context.Context, command string, cmdArgs []string) *CmdRes {
	var args []string
	if Config.Context == "" {
		args = []string{command}
	} else {
		args = []string{"--context", Config.Context, command}
	}
	args = append(args, cmdArgs...)

	res := &CmdRes{
		Cmd: fmt.Sprintf("%s %v", KubectlCmd, args),
	}

	cmd := exec.CommandContext(ctx, KubectlCmd, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = io.Writer(&res.stdout)
	cmd.Stdin = io.Reader(&res.stderr)

	startTime := time.Now()
	err := cmd.Run()
	dur := time.Since(startTime)

	res.Duration = dur

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			res.exitcode = exitError.ExitCode()
		}

		res.err = err
	}

	return res
}

// Apply applies a yaml configuration to Kubernetes cluster using kubectl
func (k *kubectl) Apply(file string) error {
	ctx, cancel := context.WithTimeout(context.Background(), kubectlActionTimeout)
	defer cancel()

	return k.RunCmd(ctx, kubectlApplyCmd, []string{
		"-f",
		file,
	}).Error()
}

// Delete deletes the provided yaml configuration in the file from the kubernetes
// cluster using kubectl.
func (k *kubectl) Delete(file string) error {
	ctx, cancel := context.WithTimeout(context.Background(), kubectlActionTimeout)
	defer cancel()

	return k.RunCmd(ctx, kubectlDeleteCmd, []string{
		"-f",
		file,
	}).Error()
}

// KubeClient represents a new client which can be used to interact with the kubernetes
// API Server.
type KubeClient struct {
	// K8s is the kubernetes interface client
	K8s kubernetes.Interface

	// APIExt is the api extension client for kubernetes.
	APIExt apiextensionsclient.Interface

	// Timeout is the default value of timeout to be used for any request we issue to the
	// API Server.
	Timeout time.Duration

	// PollInterval is the default value for polling when watching a resource using kuberentes
	// API server
	PollInterval time.Duration
}

// NewKubeClient returns a new kuberentes client wrapped inside a custom object
// which can be used to perform actions with the API servers.
func NewKubeClient(kubeconfig string) (*KubeClient, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("building config from flags failed: %s", err)
	}

	cli, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error while creating kubernetes client from config: %s", err)
	}

	apiExtCli, err := apiextensionsclient.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error while creating apiext client from config: %s", err)
	}

	return &KubeClient{
		K8s:          cli,
		APIExt:       apiExtCli,
		Timeout:      defaultRequestTimeout,
		PollInterval: pollInterval,
	}, nil
}

// WaitForCRD waits for the CRD to come up.
func (k *KubeClient) WaitForCRD(crdName string) error {
	err := wait.Poll(k.PollInterval, k.Timeout,
		func() (bool, error) {
			crd, err := k.APIExt.
				ApiextensionsV1().
				CustomResourceDefinitions().
				Get(crdName, metav1.GetOptions{})
			if err != nil {
				return false, err
			}
			for _, cond := range crd.Status.Conditions {
				switch cond.Type {
				case apiextv1.Established:
					if cond.Status == apiextv1.ConditionTrue {
						return true, err
					}
				case apiextv1.NamesAccepted:
					if cond.Status == apiextv1.ConditionFalse {
						return false, fmt.Errorf("name conflict for CRD: %s", crdName)
					}
				}
			}
			return false, err
		})

	return err
}
