package client

import (
	"os"
	"sync"

	"github.com/containerd/containerd"
)

var (
	once sync.Once
	cli  *containerd.Client
	err  error
)

func Instance() (*containerd.Client, error) {
	once.Do(func() {
		socket := os.Getenv("CONTAINERD_SOCK")
		if socket == "" {
			socket = "/run/containerd/containerd.sock"
		}
		cli, err = containerd.New(socket)
	})
	return cli, err
}
