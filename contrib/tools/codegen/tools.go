// +build tools
// To skip test by golangci-lint, as this returns an error in `go list`.

package tools

// This package imports things required by build scripts, to force `go mod` to see them as
// dependencies
//
// For now this includes k8s code-generator which will force go mod to vendor code-generator
// package this package is used by the make target `generate-k8s-client` to automatically
// generate clientset, informers and lister codes.
import (
	_ "k8s.io/code-generator"
)
