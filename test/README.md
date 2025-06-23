# Boxy Tests

This directory contains comprehensive tests for the Boxy container runtime.

## Test Structure

### `port_parse_test.go`
Tests for port parsing functionality:
- Port mapping validation (`8080:80`, `8080:80/tcp`, `8080:80/udp`)
- Protocol validation (TCP/UDP)
- Port range validation (1-65535)
- Error handling for invalid formats
- Performance benchmarks

### `cni_test.go`
Tests for CNI (Container Network Interface) functionality:
- CNI configuration validation
- Port mapping conversion
- Network namespace path generation
- Context handling

### `port_conflict_test.go`
Tests for port conflict detection:
- Available port detection
- Port-in-use detection (TCP/UDP)
- Multiple port validation
- Port availability checking

## Running Tests

### Run All Tests
```bash
cd test
go test -v .
```

### Run Specific Test File
```bash
go test -v ./port_parse_test.go
go test -v ./cni_test.go
go test -v ./port_conflict_test.go
```

### Run Benchmarks
```bash
go test -bench=.
```

### Run Tests with Coverage
```bash
go test -cover .
```

## Test Coverage

The tests cover:
- ✅ Port parsing and validation
- ✅ Protocol handling (TCP/UDP)
- ✅ Error cases and edge conditions
- ✅ CNI configuration structure
- ✅ Network namespace utilities
- ✅ Port conflict detection
- ✅ Port availability checking
- ✅ Performance benchmarks

## Adding New Tests

When adding new functionality to Boxy, please:
1. Add corresponding tests in this directory
2. Follow the existing naming conventions
3. Include both positive and negative test cases
4. Add benchmarks for performance-critical code
5. Update this README with new test descriptions

## Test Data

Test data and fixtures should be kept minimal and self-contained within the test functions to avoid external dependencies. 