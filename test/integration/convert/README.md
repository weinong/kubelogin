# Kubelogin Convert-Kubeconfig Integration Tests

This directory contains integration tests for the `kubelogin convert-kubeconfig` command that execute the actual kubelogin binary to validate end-to-end conversion functionality.

## Overview

These tests complement the existing unit tests by:
- Testing the actual binary execution and CLI argument parsing
- Validating real kubeconfig file I/O operations
- Ensuring conversion logic works end-to-end
- Testing error handling with actual command execution

## Directory Structure

```
test/integration/convert/
├── convert_test.go              # Main integration tests
├── _output/                     # Converted kubeconfigs (test-convert-with-output)
├── fixtures/
│   └── input/                   # Test kubeconfig input files (sanitized)
│       ├── devicecode-exec.yaml        # Existing exec format with devicecode
│       ├── legacy-azure-provider.yaml  # Legacy azure auth-provider format
│       ├── spn-exec.yaml               # Service principal exec format
│       ├── ropc-exec.yaml              # ROPC (Resource Owner Password Credentials) exec format
│       ├── interactive-exec.yaml       # Interactive login exec format
│       ├── workloadidentity-exec.yaml  # Workload identity exec format
│       ├── azd-exec.yaml               # Azure Developer CLI exec format
│       ├── multi-user-exec.yaml        # Multi-user/multi-cluster configuration
│       ├── large-multi-cluster.yaml    # Large kubeconfig for performance testing
│       ├── malformed-kubeconfig.yaml   # Malformed kubeconfig for error testing
│       ├── missing-required-fields.yaml # Missing required fields for validation
│       └── mixed-auth-methods.yaml     # Mixed authentication methods
├── testutils/
│   ├── binary.go               # Kubelogin binary execution helpers
│   ├── kubeconfig.go           # Kubeconfig validation utilities
│   └── sanitizer.go            # Data sanitization utilities
└── README.md                   # This file
```

## Running Tests

### Basic Integration Tests
```bash
make test-convert                # All convert integration tests
make test-convert-smoke          # Quick smoke test only
```

### Output and Inspection
```bash
make test-convert-with-output    # Save converted kubeconfigs for inspection
```

### Advanced Testing
```bash
make test-convert-mixed-auth     # Mixed authentication method tests
make test-convert-comprehensive  # All convert tests combined
```

### All Tests
```bash
make test-all                    # Unit tests + comprehensive integration tests
```

### Clean Test Output
```bash
make clean-test-output
```

### Prerequisites
- Kubelogin binary must be built first: `make kubelogin`
- Go 1.19+ required for integration test build tags

## Test Categories

### Basic Conversion Tests
- **devicecode → msi**: Convert from devicecode login to MSI
- **devicecode → azurecli**: Convert from devicecode to Azure CLI  
- **legacy → exec**: Convert legacy auth-provider to exec format

### All Authentication Methods
- **Service Principal (spn)**: Tests client certificate and secret-based authentication
- **ROPC**: Resource Owner Password Credentials authentication
- **Interactive**: Interactive browser-based authentication
- **Workload Identity**: Kubernetes workload identity authentication
- **Azure Developer CLI (azd)**: Azure Developer CLI authentication
- **MSI**: Managed Service Identity authentication
- **Azure CLI**: Azure CLI-based authentication

### Flag Combination Tests
- **client-id override**: Test --client-id flag override
- **tenant-id override**: Test --tenant-id flag override
- **environment override**: Test --environment flag (AzureUSGovernment, etc.)
- **legacy flag**: Test --legacy flag behavior
- **Multiple flags**: Test combinations of multiple flag overrides

### Complex Scenario Tests
- **Multi-user kubeconfigs**: Test conversion with multiple users and authentication methods
- **Large kubeconfigs**: Performance testing with multiple clusters and users
- **Mixed authentication**: Handle kubeconfigs with different auth methods per user

### Error and Edge Case Tests
- **Invalid login methods**: Test unsupported authentication methods
- **Missing files**: Test with non-existent kubeconfig files
- **Malformed configs**: Test with invalid YAML or missing required fields
- **Invalid flag combinations**: Test incompatible flag combinations
- **Missing required fields**: Test configs lacking mandatory authentication parameters

## Test Structure

### Main Test Function
- **TestConvertKubeconfigIntegration**: Comprehensive table-driven test with 20+ test cases
  - Covers all authentication methods, flag combinations, and real-world scenarios
  - Uses optional description field for contextual information
  - Consolidated from previous separate test functions for better maintainability

### Specialized Test Functions  
- **TestConvertKubeconfigErrors**: Error handling and edge cases
- **TestConvertSmoke**: Quick smoke test for basic functionality
- **TestConvertMixedAuthMethods**: Mixed authentication method handling

## Test Data

All test kubeconfig files are sanitized to remove sensitive information:
- **Cluster servers**: Replaced with generic test URLs
- **Certificates**: Replaced with test certificate data  
- **Names**: Use generic names like `test-cluster`, `test-user`, `test-context`
- **IDs preserved**: tenant-id, client-id, server-id kept for conversion logic testing
- **Secrets redacted**: All sensitive values replaced with test placeholders

## Adding New Tests

1. **Add fixture**: Create sanitized kubeconfig in `fixtures/input/`
2. **Add test case**: Update `TestConvertKubeconfigIntegration` with new scenario
3. **Define expectations**: Use `ConversionExpectation` to specify expected results
4. **Run tests**: Execute `make test-convert` to validate

### Example Test Case
```go
{
    name:        "Convert devicecode to interactive",
    fixture:     "devicecode-exec.yaml",
    loginMethod: "interactive",
    additionalFlags: map[string]string{
        "redirect-url": "http://localhost:8080",
    },
    expectation: testutils.ConversionExpectation{
        ExpectedLoginMethod: "interactive",
        ExpectedArgs:        []string{"--redirect-url", "http://localhost:8080"},
        ShouldHaveExec:      true,
    },
},
```

## Troubleshooting

### Binary Not Found
- Ensure `make kubelogin` has been run to build the binary
- Check that binary exists in `bin/$(OS)_$(ARCH)/kubelogin`

### Test Failures
- Use `-v` flag to see detailed test output
- Check binary execution logs in test output
- Review kubeconfig diff logs for validation failures

### Integration vs Unit Tests
- Integration tests focus on binary execution and file I/O
- Unit tests (in `pkg/`) focus on internal logic and mocking
- Both are complementary and test different aspects of the system

## CI/CD Integration

These tests are designed to run in CI/CD pipelines:
- Use build tags (`//go:build integration`) to separate from unit tests
- Have reasonable timeouts (5 minutes default)
- Provide clear error messages for debugging
- Support parallel execution with proper test isolation

## Future Enhancements

- Add tests for all authentication modes (spn, workload identity, etc.)
- Validate environment variable handling
- Test kubeconfig backup and restoration functionality
