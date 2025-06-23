package cni

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// IsRootless detects if we're running in rootless mode
func IsRootless() bool {
	return os.Geteuid() != 0
}

// createRootlessCNIConfig creates a CNI configuration optimized for rootless mode
func createRootlessCNIConfig(confFile string) error {
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
					"subnet": "10.88.0.0/16", // Different subnet for rootless
					"routes": []map[string]interface{}{
						{"dst": "0.0.0.0/0"},
					},
				},
			},
			{
				"type":         "portmap",
				"capabilities": map[string]bool{"portMappings": true},
				"snat":         true,
			},
			{
				"type": "bypass4netns",
				"capabilities": map[string]bool{
					"portMappings": true,
				},
			},
		},
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(confFile, data, 0644)
}

// NewRootlessClient creates a CNI client optimized for rootless mode
func NewRootlessClient() (*Client, error) {
	// Use user-specific CNI config directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %v", err)
	}

	confDir := filepath.Join(homeDir, ".config", "cni", "net.d")
	if err := os.MkdirAll(confDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create CNI config dir: %v", err)
	}

	// Create rootless CNI config file if it doesn't exist
	confFile := filepath.Join(confDir, "10-boxy-bridge.conflist")
	if _, err := os.Stat(confFile); os.IsNotExist(err) {
		if err := createRootlessCNIConfig(confFile); err != nil {
			return nil, fmt.Errorf("failed to create rootless CNI config: %v", err)
		}
	}

	// Check for bypass4netns plugin
	bypass4netnsPath := "/opt/cni/bin/bypass4netns"
	if _, err := os.Stat(bypass4netnsPath); os.IsNotExist(err) {
		fmt.Printf("Warning: bypass4netns plugin not found at %s\n", bypass4netnsPath)
		fmt.Printf("Rootless port forwarding may not work for privileged ports (<1024)\n")
		fmt.Printf("Install bypass4netns for better rootless support\n")
	}

	// Use standard CNI plugin directories but prefer user-local if available
	pluginDirs := []string{
		filepath.Join(homeDir, ".local", "lib", "cni"),
		"/opt/cni/bin",
		"/usr/lib/cni",
		"/usr/libexec/cni",
	}

	return createCNIClient(confDir, pluginDirs)
}

// ValidateRootlessPortMapping checks if port mappings are valid for rootless mode
func ValidateRootlessPortMapping(mappings []PortMapping) error {
	if !IsRootless() {
		return nil // No restrictions for root
	}

	for _, mapping := range mappings {
		if mapping.HostPort < 1024 {
			// Check if bypass4netns is available
			if _, err := os.Stat("/opt/cni/bin/bypass4netns"); os.IsNotExist(err) {
				return fmt.Errorf("rootless mode cannot bind to privileged port %d without bypass4netns plugin", mapping.HostPort)
			}
		}
	}
	return nil
}