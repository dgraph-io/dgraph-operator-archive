package labels

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestLabelSet(t *testing.T) {
	g := NewGomegaWithT(t)
	l := NewLabelSet()

	var tests = []struct {
		key   string
		value string
	}{
		{"newlabel.dgraph.io", "true"},
		{"testlabel.dgraph.io", "true"},
	}

	for _, test := range tests {
		l.Set(test.key, test.value)
	}

	for _, test := range tests {
		g.Expect(l.Has(test.key)).Should(BeTrue())
		g.Expect(l.Get(test.key)).To(Equal(test.value))
	}

	g.Expect(l.Has("does-not-exist")).Should(BeFalse())
	g.Expect(l.Get("does-not-exist")).To(Equal(""))

	g.Expect(l.String()).To(Equal("newlabel.dgraph.io=true,testlabel.dgraph.io=true"))

	// Test predefined labels
	var preKeyTests = []struct {
		key   K8SLabelKey
		value string
	}{
		{NameLabelKey, "dgraph"},
		{InstanceLabelKey, "dgraph-instance"},
		{ManagedByLabelKey, "dgraph-operator"},
		{ComponentLabelKey, "dgraph-test"},
	}
	l = NewLabelSet()

	for _, test := range preKeyTests {
		l.Set(string(test.key), test.value)
	}

	lNew := NewLabelSet()
	lNew.Name("dgraph")
	lNew.Instance("dgraph-instance")
	lNew.ManagedBy("dgraph-operator")
	lNew.Component("dgraph-test")

	g.Expect(l.String()).To(Equal(lNew.String()))
}
