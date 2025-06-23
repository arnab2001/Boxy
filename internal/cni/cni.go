package cni

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	gocni "github.com/containerd/go-cni"
)

// PortMapping represents a port mapping for CNI
type PortMapping struct {
	HostPort      int32  `json:"hostPort"`
	ContainerPort int32  `json:"containerPort"`
	Protocol      string `json:"protocol,omitempty"`
	HostIP        string `json:"hostIP,omitempty"`
}

// Client wraps the CNI client for Boxy
type Client struct {
	cni gocni.CNI
}

// NewClient creates a new CNI client for Boxy (auto-detects root/rootless)
func NewClient() (*Client, error) {
	if IsRootless() {
		return NewRootlessClient()
	}

	// Root mode - use system directories
	confDir := "/etc/cni/net.d"
	if err := os.MkdirAll(confDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create CNI config dir: %v", err)
	}

	confFile := filepath.Join(confDir, "10-boxy-bridge.conflist")
	if _, err := os.Stat(confFile); os.IsNotExist(err) {
		if err := createDefaultCNIConfig(confFile); err != nil {
			return nil, fmt.Errorf("failed to create CNI config: %v", err)
		}
	}

	pluginDirs := []string{"/opt/cni/bin"}
	return createCNIClient(confDir, pluginDirs)
}

// createCNIClient creates a CNI client with the given configuration
func createCNIClient(confDir string, pluginDirs []string) (*Client, error) {
	// Initialize CNI client
	cniClient, err := gocni.New(
		gocni.WithMinNetworkCount(1),
		gocni.WithPluginConfDir(confDir),
		gocni.WithPluginDir(pluginDirs),
		gocni.WithInterfacePrefix("eth"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create CNI client: %v", err)
	}

	// Load CNI configuration
	if err := cniClient.Load(gocni.WithDefaultConf); err != nil {
		return nil, fmt.Errorf("failed to load CNI config: %v", err)
	}

	return &Client{cni: cniClient}, nil
}

// SetupNetwork sets up networking for a container with port mappings
func (c *Client) SetupNetwork(ctx context.Context, containerID, netnsPath string, portMappings []PortMapping) (*gocni.Result, error) {
	// Validate port mappings for rootless mode
	if err := ValidateRootlessPortMapping(portMappings); err != nil {
		return nil, err
	}

	labels := map[string]string{
		"BOXY_CONTAINER_ID": containerID,
		"IgnoreUnknown":     "1",
	}

	// Convert port mappings to the format expected by CNI
	var cniPortMappings []gocni.PortMapping
	for _, pm := range portMappings {
		cniPortMappings = append(cniPortMappings, gocni.PortMapping{
			HostPort:      pm.HostPort,
			ContainerPort: pm.ContainerPort,
			Protocol:      pm.Protocol,
			HostIP:        pm.HostIP,
		})
	}

	// Setup network with port mappings
	result, err := c.cni.Setup(ctx, containerID, netnsPath,
		gocni.WithLabels(labels),
		gocni.WithCapabilityPortMap(cniPortMappings),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to setup network: %v", err)
	}

	return result, nil
}

// RemoveNetwork removes networking for a container
func (c *Client) RemoveNetwork(ctx context.Context, containerID, netnsPath string) error {
	labels := map[string]string{
		"BOXY_CONTAINER_ID": containerID,
		"IgnoreUnknown":     "1",
	}

	if err := c.cni.Remove(ctx, containerID, netnsPath, gocni.WithLabels(labels)); err != nil {
		return fmt.Errorf("failed to remove network: %v", err)
	}

	return nil
}

// createDefaultCNIConfig creates a default CNI configuration for Boxy (root mode)
func createDefaultCNIConfig(confFile string) error {
	config := map[string]interface{}{
		"cniVersion": "1.0.0",
		"name":       "boxy-bridge",
		"plugins": []map[string]interface{}{
			{
				"type":        "bridge",
				"bridge":      "boxy0",
				"isGateway":   true,
				"ipMasq":      true,
				"hairpinMode": true,
				"ipam": map[string]interface{}{
					"type":   "host-local",
					"subnet": "172.18.0.0/16",
					"routes": []map[string]interface{}{
						{"dst": "0.0.0.0/0"},
					},
				},
			},
			{
				"type":         "portmap",
				"capabilities": map[string]bool{"portMappings": true},
			},
		},
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(confFile, data, 0644)
} 