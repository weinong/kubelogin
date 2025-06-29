# Specification: Testing Standards and Practices

**Version:** 1.1

**Last Updated:** 2025-06-28

**Owner:** weinong

## 1. Purpose & Scope

This specification defines the testing standards, practices, and workflow for the kubelogin project. It establishes `make test` as the primary testing target and outlines comprehensive testing guidelines for Go development including integration testing for binary execution.

**Scope:**
- Full test suite execution via `make test` 
- Unit testing for internal logic and validation
- Integration testing for binary execution and end-to-end workflows
- Linting and code quality checks
- Test organization and naming conventions
- CI/CD testing requirements

## 2. Core Principles & Guidelines

### Primary Testing Target
* **`make test` is the single source of truth** for running the complete test suite
* All developers must run `make test` before submitting code
* CI/CD pipelines must use `make test` for validation
* `make test` includes both testing and linting phases

### Test Execution Hierarchy
* `make test` → Primary target (includes all below)
  * `make lint` → golangci-lint checks via `.golangci-lint` tool
  * `go test -race -coverprofile=coverage.txt -covermode=atomic ./pkg/...` → Unit tests with race detection and coverage
* `make test-all` → Complete test suite (unit + integration)
  * `make test` → Unit tests and linting
  * `make test-convert` → Integration tests for convert-kubeconfig binary
* `make test-convert-smoke` → Quick smoke test for integration functionality
* `make test-convert-with-output` → Integration tests with output directory for manual inspection

### Coverage and Quality Requirements
* Tests run with race detection (`-race`) to catch concurrency issues
* Coverage report generated (`coverage.txt`) with atomic mode
* All linting rules must pass before tests execute
* CI/CD pipelines must achieve the same results as local `make test`

### Testing Standards
* All new code must include corresponding tests
* **Unit tests** must be co-located with source code (`*_test.go` files)
* **Integration tests** are organized in `test/integration/` directory structure
* Test functions follow naming: `TestFunctionName_Scenario_ExpectedResult`
* Use table-driven tests for multiple test cases
* Mock external dependencies using interfaces
* Integration tests use build tags (`//go:build integration`) for selective execution

### Integration Testing Requirements
* Binary-based integration tests validate end-to-end functionality
* Tests execute the actual compiled kubelogin binary
* Sanitized test fixtures protect sensitive data
* Optimal file placement minimizes I/O operations during test execution
* Output directory support enables manual inspection of conversion results

## 3. Rationale & Context

### Why `make test` as Primary Target
- **Consistency**: Single command for all developers regardless of environment
- **Completeness**: Ensures both testing and code quality checks are performed
- **CI/CD Integration**: Simplifies automation and build pipelines
- **Documentation**: Clear, discoverable entry point for testing

### Testing Philosophy
The kubelogin project follows a comprehensive testing approach that balances thorough coverage with maintainable test code. We prioritize:
- **Type Safety**: Using the new options pattern ensures compile-time validation
- **Integration Confidence**: Testing real authentication flows where possible
- **Binary Validation**: End-to-end testing of CLI commands and flag processing
- **Mocking Strategy**: Clean interfaces for external dependencies (Azure AD, Kubernetes)
- **Performance**: Fast test execution for developer productivity
- **Manual Inspection**: Output directory support for debugging conversion results

### Integration Testing Strategy
Integration tests complement unit tests by validating binary execution and real file I/O:
- **Binary Execution**: Tests execute the actual kubelogin binary, not mocked functions
- **Fixture Management**: Sanitized kubeconfig files remove sensitive data while preserving test functionality
- **Optimal I/O**: Smart file placement reduces redundant copy operations
- **Build Tags**: `//go:build integration` enables selective test execution
- **Error Validation**: Tests verify both success and failure scenarios with actual CLI error handling

## 4. Examples

### Good Example (Do)

```makefile
# Updated Makefile implementation with integration testing
test: lint
	go test -race -coverprofile=coverage.txt -covermode=atomic ./pkg/...

test-convert: $(TARGET)
	go test ./test/integration/convert/... -tags=integration -timeout=5m

test-convert-smoke: $(TARGET)
	go test ./test/integration/convert/... -tags=integration -run=TestConvertSmoke

test-convert-with-output: $(TARGET)
	@mkdir -p test/integration/convert/_output
	KUBELOGIN_TEST_OUTPUT_DIR=$(PWD)/test/integration/convert/_output go test ./test/integration/convert/... -tags=integration -timeout=5m

test-all: test test-convert

clean-test-output:
	-rm -rf test/integration/convert/_output
```

```bash
# Proper way to run tests
$ make test                    # Unit tests + linting
$ make test-all               # Complete test suite (unit + integration)
$ make test-convert           # Integration tests only
$ make test-convert-with-output # Integration tests with manual inspection files
$ make test-convert-smoke     # Quick integration smoke test
```

```go
// Unit test example (existing pattern)
func TestConvert_ValidateOptions(t *testing.T) {
	tests := []struct {
		name        string
		setupOpts   func() *converter.RawOptions
		expectError bool
		errorMsg    string
	}{
		{
			name: "ValidOptions_ShouldPass",
			setupOpts: func() *converter.RawOptions {
				opts := converter.NewRawOptions()
				opts.SetContext("test-context")
				return opts
			},
			expectError: false,
		},
		{
			name: "MissingServerID_ShouldFail",
			setupOpts: func() *converter.RawOptions {
				opts := converter.NewRawOptions()
				// Missing required ServerID
				return opts
			},
			expectError: true,
			errorMsg:    "server-id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := tt.setupOpts()
			validated, err := opts.Validate()
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, validated)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, validated)
			}
		})
	}
}
```

```go
// Integration test example (new pattern)
//go:build integration

package convert

import (
	"path/filepath"
	"testing"
	"github.com/Azure/kubelogin/test/integration/convert/testutils"
)

func TestConvertKubeconfigIntegration(t *testing.T) {
	binary := testutils.NewKubeloginBinary(t)
	validator := testutils.NewKubeconfigValidator(t)

	tests := []struct {
		name        string
		fixture     string
		loginMethod string
		expectation testutils.ConversionExpectation
	}{
		{
			name:        "Convert devicecode exec to MSI",
			fixture:     "devicecode-exec.yaml",
			loginMethod: "msi",
			expectation: testutils.ConversionExpectation{
				ExpectedCommand:     "kubelogin",
				ExpectedLoginMethod: "msi",
				ExpectedServerID:    "6dae42f8-4368-4678-94ff-3960e28e3630",
				ShouldHaveExec:      true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Optimal file placement based on output directory setting
			fixturePath := filepath.Join("fixtures", "input", tt.fixture)
			kubeconfig := testutils.PrepareKubeconfigForTest(t, fixturePath, tt.name)

			// Execute actual binary with real CLI argument parsing
			stdout, stderr, err := binary.ConvertKubeconfigWithFlags(
				kubeconfig, tt.loginMethod, nil)

			if err != nil {
				t.Fatalf("Conversion failed: %v. STDERR: %s", err, stderr)
			}

			// Validate real file I/O results
			validator.ValidateConversion(fixturePath, kubeconfig, tt.expectation)
		})
	}
}
```

### Bad Example (Don't / Avoid)

```bash
# Don't: Direct test execution without make target
$ go test ./pkg/internal/converter  # Bypasses linting

# Don't: Running integration tests without building binary first
$ go test ./test/integration/convert/... -tags=integration  # Binary may not exist

# Don't: Ignoring integration tests in CI/CD
$ make test  # Only unit tests, missing integration coverage
```

```go
// Don't: Poor test naming
func TestConvert(t *testing.T) { /* unclear what's being tested */ }

// Don't: No test isolation
func TestMultipleScenarios(t *testing.T) {
	// Testing multiple unrelated scenarios in one function
	// Makes debugging failures difficult
}

// Don't: Accessing private fields (breaks encapsulation)
func TestOptions(t *testing.T) {
	opts := &CompletedOptions{}
	opts.validated.tokenOptions = &token.Options{} // Direct access to private field
}

// Don't: Integration tests without build tags
package convert
import "testing"

func TestBinaryExecution(t *testing.T) {
	// Missing: //go:build integration
	// This will run with unit tests, slowing them down
}

// Don't: Inefficient file operations in integration tests
func TestConversion(t *testing.T) {
	// Copy to temp
	temp := copyToTemp(fixture)
	// Convert in temp
	convert(temp)
	// Copy to output  ← Redundant operation!
	copyToOutput(temp, outputDir)
}

// Don't: Hardcoded sensitive data in test fixtures
apiVersion: v1
clusters:
- cluster:
    server: https://real-production-cluster.example.com  # Don't expose real FQDNs
    certificate-authority-data: LS0...Real-Cert-Data...  # Don't include real certificates
```

## 5. Related Specifications / Further Reading

- [Integration Testing Documentation](../../../test/integration/convert/README.md) - Convert-kubeconfig integration test setup and usage
- [CLI Flags Specification](../cli-flags/main.spec.md) - For testing flag validation
- [Go Testing Documentation](https://golang.org/pkg/testing/) - Official Go testing package
- [Go Build Constraints](https://pkg.go.dev/go/build#hdr-Build_Constraints) - Build tags for integration tests
- [Testify Library](https://github.com/stretchr/testify) - Assertion library used in project
- [golangci-lint](https://golangci-lint.run/) - Linter configuration and rules

## 6. Keywords

testing, make test, unit tests, integration tests, binary execution, table-driven tests, build tags, lint, go test, mocking, coverage, CI/CD, code quality, validation, kubeconfig conversion, fixture sanitization, file I/O optimization, output directory, manual inspection