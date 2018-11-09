// +build tools

package tools

import (
	// Pull in the Kubernetes code generator
	_ "k8s.io/code-generator/cmd/client-gen"
)
