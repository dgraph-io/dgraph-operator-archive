package k8s

import (
	"flag"
	"fmt"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	"github.com/spf13/viper"
)

var (
	client *KubeClient
)

func init() {
	viper.SetEnvPrefix("DGRAPH")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	flag.StringVar(&Config.Kubeconfig, "kubeconfig", defaultKubeconfigPath, "Path to the kubeconfig file")
	viper.BindEnv("kubeconfig")
	flag.StringVar(&Config.OperatorImage, "operator-image", defaultOperatorImage, "Image to use for the operator.")
	viper.BindEnv("operator-image")
	flag.StringVar(&Config.Context, "context", "", "Context to use with kubectl command line utility")
	viper.BindEnv("context")

	// We don't do a flag parse here, as ginkgo does that automatically for us.
}

// Called before running anything in the test suite
var _ = BeforeSuite(func() {
	var err error

	if _, err = os.Stat(Config.Kubeconfig); err != nil {
		Fail(fmt.Sprintf("error while opening kubeconfig file: %s", err))
	}

	client, err = NewKubeClient(Config.Kubeconfig)
	if err != nil {
		Fail(fmt.Sprintf("error while creating kube client: %s", err))
	}
})
