package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// PortMapping matches the CNI portmap plugin's expected structure
// https://www.cni.dev/plugins/current/meta/portmap/
type PortMapping struct {
	HostPort      int32  `json:"hostPort"`
	ContainerPort int32  `json:"containerPort"`
	Protocol      string `json:"protocol,omitempty"`
	HostIP        string `json:"hostIP,omitempty"`
}

// ParsePorts parses -p/--publish flags into PortMapping structs
func ParsePorts(flags []string) ([]PortMapping, error) {
	var mappings []PortMapping
	for _, flag := range flags {
		proto := "tcp" // default protocol
		hostPortStr := ""
		contPortProto := ""

		parts := strings.SplitN(flag, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid port mapping (expected HOST:CONT): %s", flag)
		}
		hostPortStr = parts[0]
		contPortProto = parts[1]

		contParts := strings.SplitN(contPortProto, "/", 2)
		contPortStr := contParts[0]
		if len(contParts) == 2 {
			proto = strings.ToLower(contParts[1])
			if proto != "tcp" && proto != "udp" {
				return nil, fmt.Errorf("unsupported protocol: %s", proto)
			}
		}

		hostPort, err1 := parsePortNum(hostPortStr)
		contPort, err2 := parsePortNum(contPortStr)
		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("invalid port numbers in: %s", flag)
		}

		mappings = append(mappings, PortMapping{
			HostPort:      hostPort,
			ContainerPort: contPort,
			Protocol:      proto,
		})
	}
	return mappings, nil
}

// parsePortNum parses a string port number
func parsePortNum(s string) (int32, error) {
	p, err := strconv.Atoi(s)
	if err != nil || p < 1 || p > 65535 {
		return 0, fmt.Errorf("invalid port: %s", s)
	}
	return int32(p), nil
}

// checkPortConflicts checks if any of the host ports are already in use
func checkPortConflicts(mappings []PortMapping) error {
	for _, mapping := range mappings {
		if err := isPortAvailable(mapping.HostPort, mapping.Protocol); err != nil {
			return fmt.Errorf("port %d/%s is already in use: %v", mapping.HostPort, mapping.Protocol, err)
		}
	}
	return nil
}

// isPortAvailable checks if a port is available for binding
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