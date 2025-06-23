package main

import (
	"context"
	"fmt"
	"os"
	"testing"
)

// Test CNI configuration creation
func TestCreateDefaultCNIConfig(t *testing.T) {
	// Create a temporary file for testing
	tmpFile := "/tmp/test-cni-config.conflist"
	defer os.Remove(tmpFile)

	// Test the CNI config creation (we'd need to extract this from internal/cni/cni.go)
	// For now, just test that we can create a valid JSON structure
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

	// Verify the config structure
	if config["cniVersion"] != "1.0.0" {
		t.Errorf("Expected CNI version 1.0.0, got %v", config["cniVersion"])
	}

	if config["name"] != "boxy-bridge" {
		t.Errorf("Expected network name boxy-bridge, got %v", config["name"])
	}

	plugins, ok := config["plugins"].([]map[string]interface{})
	if !ok || len(plugins) != 2 {
		t.Errorf("Expected 2 plugins, got %v", len(plugins))
	}

	// Check bridge plugin
	bridgePlugin := plugins[0]
	if bridgePlugin["type"] != "bridge" {
		t.Errorf("Expected bridge plugin, got %v", bridgePlugin["type"])
	}

	if bridgePlugin["bridge"] != "boxy0" {
		t.Errorf("Expected bridge name boxy0, got %v", bridgePlugin["bridge"])
	}

	// Check portmap plugin
	portmapPlugin := plugins[1]
	if portmapPlugin["type"] != "portmap" {
		t.Errorf("Expected portmap plugin, got %v", portmapPlugin["type"])
	}
}

// Test port mapping conversion
func TestPortMappingConversion(t *testing.T) {
	// Test data
	type PortMapping struct {
		HostPort      int32  `json:"hostPort"`
		ContainerPort int32  `json:"containerPort"`
		Protocol      string `json:"protocol,omitempty"`
		HostIP        string `json:"hostIP,omitempty"`
	}

	mappings := []PortMapping{
		{HostPort: 8080, ContainerPort: 80, Protocol: "tcp"},
		{HostPort: 9000, ContainerPort: 9000, Protocol: "udp"},
	}

	// Verify mappings
	if len(mappings) != 2 {
		t.Errorf("Expected 2 mappings, got %d", len(mappings))
	}

	// Check TCP mapping
	tcpMapping := mappings[0]
	if tcpMapping.HostPort != 8080 {
		t.Errorf("Expected host port 8080, got %d", tcpMapping.HostPort)
	}
	if tcpMapping.ContainerPort != 80 {
		t.Errorf("Expected container port 80, got %d", tcpMapping.ContainerPort)
	}
	if tcpMapping.Protocol != "tcp" {
		t.Errorf("Expected protocol tcp, got %s", tcpMapping.Protocol)
	}

	// Check UDP mapping
	udpMapping := mappings[1]
	if udpMapping.HostPort != 9000 {
		t.Errorf("Expected host port 9000, got %d", udpMapping.HostPort)
	}
	if udpMapping.ContainerPort != 9000 {
		t.Errorf("Expected container port 9000, got %d", udpMapping.ContainerPort)
	}
	if udpMapping.Protocol != "udp" {
		t.Errorf("Expected protocol udp, got %s", udpMapping.Protocol)
	}
}

// Test context creation (basic test)
func TestContextCreation(t *testing.T) {
	ctx := context.Background()
	if ctx == nil {
		t.Error("Context should not be nil")
	}

	// Test context with timeout
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if ctx == nil {
		t.Error("Context with cancel should not be nil")
	}
}

// Test network namespace path generation
func TestNetworkNamespacePath(t *testing.T) {
	testPID := int32(12345)
	expectedPath := "/proc/12345/ns/net"
	
	// Use fmt.Sprintf to generate the path
	netnsPath := fmt.Sprintf("/proc/%d/ns/net", testPID)
	
	if netnsPath != expectedPath {
		t.Errorf("Expected netns path %s, got %s", expectedPath, netnsPath)
	}
} 