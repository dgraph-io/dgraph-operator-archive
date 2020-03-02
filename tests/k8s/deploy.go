package k8s

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("K8sDeployTest", func() {

	var (
		crdYamlPath string = "./../contrib/crd/dgraph.io_dgraphclusters.yaml"
	)

	Context("Install", func() {
		var err error

		It("Test operator deployment with manual CRD", func() {
			By("Creating CRD")
			err = Kubectl.Apply(crdYamlPath)
			Expect(err).NotTo(HaveOccurred())

			By("Waiting for CRD to come up.")
			err = client.WaitForCRD(dgraphClusterCRD)
			Expect(err).NotTo(HaveOccurred())

			By("Setting up RBAC for operator")
			err = createOperatorRBAC()
			Expect(err).NotTo(HaveOccurred())

			By("Creating operator deployment with setup service account")
			deployment := operatorDeployment()
			res, err := client.K8s.
				AppsV1().Deployments(v1.NamespaceDefault).
				Create(deployment)
			Expect(err).NotTo(HaveOccurred())

			p := Kubectl.RunCmd(context.TODO(), "describe", []string{"deployments/dgraph-operator"})
			fmt.Println(p.GetStdout(), p.GetStderr())

			By("Deleting RBAC setup for operator")
			err = deleteOperatorRBAC()
			Expect(err).NotTo(HaveOccurred())

			By("Deleting operator deployment")
			err = client.K8s.
				AppsV1().Deployments(v1.NamespaceDefault).
				Delete(res.GetObjectMeta().GetName(), &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())

			By("Tearing down dgraph.io CRDs")
			Kubectl.Delete(crdYamlPath)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
