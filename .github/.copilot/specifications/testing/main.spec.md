# Specification: Testing Standards and Practices

**Version:** 1.0

**Last Updated:** 2025-06-28

**Owner:** weinong

## 1. Purpose & Scope

This specification defines the testing standards, practices, and workflow for the kubelogin project. It establishes `make test` as the primary testing target and outlines comprehensive testing guidelines for Go development.

**Scope:**
- Full test suite execution via `make test`
- Linting and code quality checks
- Unit and integration testing patterns
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
  * `go test -race -coverprofile=coverage.txt -covermode=atomic ./...` → Full test suite with race detection and coverage

### Coverage and Quality Requirements
* Tests run with race detection (`-race`) to catch concurrency issues
* Coverage report generated (`coverage.txt`) with atomic mode
* All linting rules must pass before tests execute
* CI/CD pipelines must achieve the same results as local `make test`

### Testing Standards
* All new code must include corresponding tests
* Tests must be co-located with source code (`*_test.go` files)
* Test functions follow naming: `TestFunctionName_Scenario_ExpectedResult`
* Use table-driven tests for multiple test cases
* Mock external dependencies using interfaces

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
- **Mocking Strategy**: Clean interfaces for external dependencies (Azure AD, Kubernetes)
- **Performance**: Fast test execution for developer productivity

## 4. Examples

### Good Example (Do)

```makefile
# Current Makefile implementation
test: lint
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run
```

```bash
# Proper way to run tests
$ make test           # Runs lint + tests with coverage
$ make lint          # Runs only linting (subset of make test)
```

```go
// Table-driven test example
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

### Bad Example (Don't / Avoid)

```go
// Don't: Direct test execution without make target
$ go test ./pkg/internal/converter  # Bypasses linting

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
```

## 5. Related Specifications / Further Reading

- [CLI Flags Specification](../cli-flags/main.spec.md) - For testing flag validation
- [Go Testing Documentation](https://golang.org/pkg/testing/) - Official Go testing package
- [Testify Library](https://github.com/stretchr/testify) - Assertion library used in project
- [golangci-lint](https://golangci-lint.run/) - Linter configuration and rules

## 6. Keywords

testing, make test, lint, go test, unit tests, integration tests, table-driven tests, mocking, coverage, CI/CD, code quality, validation