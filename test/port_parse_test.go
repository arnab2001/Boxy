package main

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// PortMapping represents a port mapping from host to container
// This is a copy of the struct from cmd/boxy/port_parse.go for testing
type PortMapping struct {
	HostPort      int32  `json:"hostPort"`
	ContainerPort int32  `json:"containerPort"`
	Protocol      string `json:"protocol,omitempty"`
	HostIP        string `json:"hostIP,omitempty"`
}

// ParsePorts parses port mapping flags like "8080:80" or "8080:80/tcp"
// This is a copy of the function from cmd/boxy/port_parse.go for testing
func ParsePorts(portFlags []string) ([]PortMapping, error) {
	if len(portFlags) == 0 {
		return nil, nil
	}

	var mappings []PortMapping
	for _, flag := range portFlags {
		if flag == "" {
			return nil, fmt.Errorf("empty port mapping")
		}

		// Split by colon to get host:container
		parts := strings.SplitN(flag, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid port mapping format: %s (expected HOST:CONTAINER)", flag)
		}

		hostPortStr := parts[0]
		containerPart := parts[1]

		// Parse host port
		hostPort, err := parsePortNum(hostPortStr)
		if err != nil {
			return nil, fmt.Errorf("invalid host port: %v", err)
		}

		// Check if container part has protocol (e.g., "80/tcp")
		var containerPortStr, protocol string
		if strings.Contains(containerPart, "/") {
			protocolParts := strings.SplitN(containerPart, "/", 2)
			if len(protocolParts) != 2 {
				return nil, fmt.Errorf("invalid container port/protocol format: %s", containerPart)
			}
			containerPortStr = protocolParts[0]
			protocol = strings.ToLower(protocolParts[1])
		} else {
			containerPortStr = containerPart
			protocol = "tcp" // default protocol
		}

		// Validate protocol
		if protocol != "tcp" && protocol != "udp" {
			return nil, fmt.Errorf("invalid protocol: %s (must be tcp or udp)", protocol)
		}

		// Parse container port
		containerPort, err := parsePortNum(containerPortStr)
		if err != nil {
			return nil, fmt.Errorf("invalid container port: %v", err)
		}

		mappings = append(mappings, PortMapping{
			HostPort:      hostPort,
			ContainerPort: containerPort,
			Protocol:      protocol,
			HostIP:        "", // Default to all interfaces
		})
	}

	return mappings, nil
}

// parsePortNum parses and validates a port number string
// This is a copy of the function from cmd/boxy/port_parse.go for testing
func parsePortNum(portStr string) (int32, error) {
	if portStr == "" {
		return 0, fmt.Errorf("empty port number")
	}

	port, err := strconv.ParseInt(portStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid port number: %s", portStr)
	}

	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("port number out of range: %d (must be 1-65535)", port)
	}

	return int32(port), nil
}

func TestParsePorts(t *testing.T) {
	tests := []struct {
		name     string
		flags    []string
		expected []PortMapping
		wantErr  bool
	}{
		{
			name:  "single port mapping tcp",
			flags: []string{"8080:80"},
			expected: []PortMapping{
				{HostPort: 8080, ContainerPort: 80, Protocol: "tcp"},
			},
			wantErr: false,
		},
		{
			name:  "single port mapping with explicit tcp",
			flags: []string{"8080:80/tcp"},
			expected: []PortMapping{
				{HostPort: 8080, ContainerPort: 80, Protocol: "tcp"},
			},
			wantErr: false,
		},
		{
			name:  "single port mapping udp",
			flags: []string{"8080:80/udp"},
			expected: []PortMapping{
				{HostPort: 8080, ContainerPort: 80, Protocol: "udp"},
			},
			wantErr: false,
		},
		{
			name:  "multiple port mappings",
			flags: []string{"8080:80", "9000:9000/udp", "3000:3000/tcp"},
			expected: []PortMapping{
				{HostPort: 8080, ContainerPort: 80, Protocol: "tcp"},
				{HostPort: 9000, ContainerPort: 9000, Protocol: "udp"},
				{HostPort: 3000, ContainerPort: 3000, Protocol: "tcp"},
			},
			wantErr: false,
		},
		{
			name:  "high port numbers",
			flags: []string{"65535:65534"},
			expected: []PortMapping{
				{HostPort: 65535, ContainerPort: 65534, Protocol: "tcp"},
			},
			wantErr: false,
		},
		{
			name:  "low port numbers",
			flags: []string{"1:1"},
			expected: []PortMapping{
				{HostPort: 1, ContainerPort: 1, Protocol: "tcp"},
			},
			wantErr: false,
		},
		{
			name:     "empty flags",
			flags:    []string{},
			expected: nil,
			wantErr:  false,
		},
		{
			name:    "invalid format - missing colon",
			flags:   []string{"8080"},
			wantErr: true,
		},
		{
			name:    "invalid format - empty string",
			flags:   []string{""},
			wantErr: true,
		},
		{
			name:    "invalid format - multiple colons",
			flags:   []string{"8080:80:90"},
			wantErr: true,
		},
		{
			name:    "invalid host port - non-numeric",
			flags:   []string{"abc:80"},
			wantErr: true,
		},
		{
			name:    "invalid container port - non-numeric",
			flags:   []string{"8080:abc"},
			wantErr: true,
		},
		{
			name:    "invalid host port - zero",
			flags:   []string{"0:80"},
			wantErr: true,
		},
		{
			name:    "invalid container port - zero",
			flags:   []string{"8080:0"},
			wantErr: true,
		},
		{
			name:    "invalid host port - too high",
			flags:   []string{"65536:80"},
			wantErr: true,
		},
		{
			name:    "invalid container port - too high",
			flags:   []string{"8080:65536"},
			wantErr: true,
		},
		{
			name:    "invalid protocol",
			flags:   []string{"8080:80/http"},
			wantErr: true,
		},
		{
			name:    "protocol case insensitive",
			flags:   []string{"8080:80/TCP", "9000:9000/UDP"},
			expected: []PortMapping{
				{HostPort: 8080, ContainerPort: 80, Protocol: "tcp"},
				{HostPort: 9000, ContainerPort: 9000, Protocol: "udp"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParsePorts(tt.flags)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParsePorts() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("ParsePorts() unexpected error: %v", err)
				return
			}
			
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParsePorts() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestParsePortNum(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int32
		wantErr  bool
	}{
		{
			name:     "valid port 80",
			input:    "80",
			expected: 80,
			wantErr:  false,
		},
		{
			name:     "valid port 8080",
			input:    "8080",
			expected: 8080,
			wantErr:  false,
		},
		{
			name:     "valid port 1",
			input:    "1",
			expected: 1,
			wantErr:  false,
		},
		{
			name:     "valid port 65535",
			input:    "65535",
			expected: 65535,
			wantErr:  false,
		},
		{
			name:    "invalid port 0",
			input:   "0",
			wantErr: true,
		},
		{
			name:    "invalid port 65536",
			input:   "65536",
			wantErr: true,
		},
		{
			name:    "invalid port negative",
			input:   "-1",
			wantErr: true,
		},
		{
			name:    "invalid port non-numeric",
			input:   "abc",
			wantErr: true,
		},
		{
			name:    "invalid port empty",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid port with spaces",
			input:   " 80 ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parsePortNum(tt.input)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("parsePortNum() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("parsePortNum() unexpected error: %v", err)
				return
			}
			
			if result != tt.expected {
				t.Errorf("parsePortNum() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// Benchmark tests for performance
func BenchmarkParsePorts(b *testing.B) {
	flags := []string{"8080:80", "9000:9000/udp", "3000:3000/tcp", "443:443", "22:22/tcp"}
	
	for i := 0; i < b.N; i++ {
		_, err := ParsePorts(flags)
		if err != nil {
			b.Fatalf("ParsePorts() error: %v", err)
		}
	}
}

func BenchmarkParsePortNum(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := parsePortNum("8080")
		if err != nil {
			b.Fatalf("parsePortNum() error: %v", err)
		}
	}
} 