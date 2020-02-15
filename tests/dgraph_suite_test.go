package dgraph_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDgraphOperator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dgraph operator test Suite")
}
