package oci

import (
	"context"

	"github.com/containerd/containerd/containers"
	"github.com/containerd/containerd/oci"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

func WithImageConfig(img oci.Image) oci.SpecOpts {
	return oci.WithImageConfig(img)
}

func DefaultNamespaces() oci.SpecOpts {
	return func(_ context.Context, _ oci.Client, _ *containers.Container, s *specs.Spec) error {
		// later i will tweak net / uts / ipc here
		return nil
	}
}
