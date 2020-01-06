package labels

import (
	"fmt"
	"sort"
	"strings"
)

// K8SlabelKey is the commonly used labels as kubernetes docs defines.
// https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/
type K8SLabelKey string

var (
	// NameLabelKey is the name of the application.
	NameLabelKey K8SLabelKey = "app.kubernetes.io/name"

	// InstanceLabelKey is a unique name identifying the instance of an application.
	InstanceLabelKey K8SLabelKey = "app.kubernetes.io/instance"

	// ManagedByLabelKey represents the tool being used to manage the operation of an application.
	ManagedByLabelKey K8SLabelKey = "app.kubernetes.io/managed-by"

	// ComponentLabelKey is the component within the architecture
	ComponentLabelKey K8SLabelKey = "app.kubernetes.io/component"
)

// Labels is the standard type to manage labels for the operator.
type Labels map[string]string

// NewLabels returns a new Labels type to which type can be associated.
func NewLabelSet() Labels {
	return make(map[string]string)
}

// Set a label value in the label set
func (l Labels) Set(label, value string) {
	l[label] = value
}

// Has checks if the provided label is present in the label set.
func (l Labels) Has(label string) bool {
	_, exists := l[label]
	return exists
}

// Get returns the value of the requested label.
func (l Labels) Get(label string) string {
	return l[label]
}

// String returns all labels listed as a human readable string.
func (l Labels) String() string {
	labelsList := make([]string, 0, len(l))
	for key, value := range l {
		labelsList = append(labelsList, fmt.Sprintf("%s=%s", key, value))
	}

	sort.StringSlice(labelsList).Sort()
	return strings.Join(labelsList, ",")
}

// Predefined function to set kubernetes common labels in the LabelSet.

// Name sets the NameLabelKey in the label set.
func (l Labels) Name(value string) Labels {
	l[string(NameLabelKey)] = value
	return l
}

// Instance sets the InstanceLabelKey in the lables set.
func (l Labels) Instance(value string) Labels {
	l[string(InstanceLabelKey)] = value
	return l
}

// ManagedBy sets the ManagedByLabelKey in the labels set.
func (l Labels) ManagedBy(value string) Labels {
	l[string(ManagedByLabelKey)] = value
	return l
}

// Component sets the ComponentLabelKey in the labels set.
func (l Labels) Component(value string) Labels {
	l[string(ComponentLabelKey)] = value
	return l
}
