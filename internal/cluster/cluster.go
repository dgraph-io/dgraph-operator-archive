package cluster

import (
	"github.com/dgraph-io/dgraph-operator/api/v1alpha1"
)

type State struct {
	cluster v1alpha1.DgraphCluster
	status  v1alpha1.DgraphClusterStatus
}

// NewState creates a new reconcile state based on the given cluster
func NewState(c v1alpha1.DgraphCluster) (*State, error) {
	status := *c.Status.DeepCopy()
	// reset the phase to an empty string so that we do not report an outdated phase given that certain phases are
	// stickier than others (eg. invalid)
	status.Phase = ""
	return &State{
		cluster: c,
		status:  status,
	}, nil
}
