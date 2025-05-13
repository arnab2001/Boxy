package client

import (
	"context"

	"github.com/containerd/containerd/namespaces"
)

// Default returns a context pre-populated with our namespace.
func Default() context.Context {
	return namespaces.WithNamespace(context.Background(), "boxy")
}
