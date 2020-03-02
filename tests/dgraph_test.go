package dgraph_test

import (
	_ "github.com/dgraph-io/dgraph-operator/tests/k8s"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DgraphOperator", func() {
	It("Sample Test", func() {
		Expect("test").To(Equal("test"))
	})
})
