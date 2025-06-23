package main

import (
	"fmt"
	"net"
	"testing"
	"time"
)

// Test port conflict detection
func TestPortConflictDetection(t *testing.T) {
	// Test available port
	t.Run("available_port", func(t *testing.T) {
		mappings := []PortMapping{
			{HostPort: 19999, ContainerPort: 80, Protocol: "tcp"}, // Use high port unlikely to be in use
		}
		
		err := checkPortConflicts(mappings)
		if err != nil {
			t.Errorf("Expected no conflict for available port, got: %v", err)
		}
	})

	// Test port in use
	t.Run("port_in_use", func(t *testing.T) {
		// Start a listener on a port
		listener, err := net.Listen("tcp", ":18888")
		if err != nil {
			t.Fatalf("Failed to start test listener: %v", err)
		}
		defer listener.Close()

		mappings := []PortMapping{
			{HostPort: 18888, ContainerPort: 80, Protocol: "tcp"},
		}
		
		err = checkPortConflicts(mappings)
		if err == nil {
			t.Error("Expected conflict for port in use, but got none")
		}
	})

	// Test UDP port
	t.Run("udp_port_available", func(t *testing.T) {
		mappings := []PortMapping{
			{HostPort: 19998, ContainerPort: 53, Protocol: "udp"},
		}
		
		err := checkPortConflicts(mappings)
		if err != nil {
			t.Errorf("Expected no conflict for available UDP port, got: %v", err)
		}
	})

	// Test UDP port in use
	t.Run("udp_port_in_use", func(t *testing.T) {
		// Start a UDP listener
		conn, err := net.ListenPacket("udp", ":18887")
		if err != nil {
			t.Fatalf("Failed to start test UDP listener: %v", err)
		}
		defer conn.Close()

		mappings := []PortMapping{
			{HostPort: 18887, ContainerPort: 53, Protocol: "udp"},
		}
		
		err = checkPortConflicts(mappings)
		if err == nil {
			t.Error("Expected conflict for UDP port in use, but got none")
		}
	})

	// Test multiple ports
	t.Run("multiple_ports_available", func(t *testing.T) {
		mappings := []PortMapping{
			{HostPort: 19997, ContainerPort: 80, Protocol: "tcp"},
			{HostPort: 19996, ContainerPort: 443, Protocol: "tcp"},
			{HostPort: 19995, ContainerPort: 53, Protocol: "udp"},
		}
		
		err := checkPortConflicts(mappings)
		if err != nil {
			t.Errorf("Expected no conflict for available ports, got: %v", err)
		}
	})
}

// Test port availability checking
func TestIsPortAvailable(t *testing.T) {
	tests := []struct {
		name     string
		port     int32
		protocol string
		setup    func() net.Listener // Setup function to occupy the port
		wantErr  bool
	}{
		{
			name:     "tcp_port_available",
			port:     19994,
			protocol: "tcp",
			setup:    nil,
			wantErr:  false,
		},
		{
			name:     "udp_port_available", 
			port:     19993,
			protocol: "udp",
			setup:    nil,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var listener net.Listener
			if tt.setup != nil {
				listener = tt.setup()
				defer listener.Close()
			}

			err := isPortAvailable(tt.port, tt.protocol)
			if (err != nil) != tt.wantErr {
				t.Errorf("isPortAvailable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper functions for testing
func checkPortConflicts(mappings []PortMapping) error {
	for _, mapping := range mappings {
		if err := isPortAvailable(mapping.HostPort, mapping.Protocol); err != nil {
			return fmt.Errorf("port %d/%s is already in use: %v", mapping.HostPort, mapping.Protocol, err)
		}
	}
	return nil
}

func isPortAvailable(port int32, protocol string) error {
	address := fmt.Sprintf(":%d", port)
	
	if protocol == "tcp" {
		listener, err := net.Listen("tcp", address)
		if err != nil {
			return err
		}
		listener.Close()
	} else if protocol == "udp" {
		conn, err := net.ListenPacket("udp", address)
		if err != nil {
			return err
		}
		conn.Close()
	}
	
	// Small delay to ensure the port is fully released
	time.Sleep(10 * time.Millisecond)
	return nil
} 