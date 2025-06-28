# Specification: Enhanced Options Pattern with Private Encapsulation

**Version:** 2.0

**Last Updated:** 2025-06-28

**Owner:** weinong

**Status:** ✅ IMPLEMENTED

## 1. Purpose & Scope

This document defines a structured process for handling command-line options by separating their construction into three clear phases with enhanced type safety and encapsulation.

1. **RawOptions**: Hold all raw (unvalidated) input values (e.g., from CLI flags).
2. **Validate()**: Validate the raw options, returning a **ValidatedOptions** object with private encapsulated data if checks pass.
3. **Complete()**: Complete the validated options to produce a **CompletedOptions** object, sharing validated data immutably.

Each phase enforces order and type safety:
* You must call Validate() before Complete().
* Validated data is privately encapsulated and cannot be modified externally.
* Only fully constructed CompletedOptions are ready for use.

## 2. Architecture Overview

### Enhanced Encapsulation Pattern

```go
// Private validated data - cannot be instantiated outside package
type validatedData struct {
    configFlags    genericclioptions.RESTClientGetter
    flags          *pflag.FlagSet
    tokenOptions   *token.Options
    context        string
    azureConfigDir string
}

// ValidatedOptions wraps validated data with controlled access
type ValidatedOptions struct {
    validated *validatedData  // Private pointer - no external access
}

// CompletedOptions shares validated data immutably
type CompletedOptions struct {
    validated *validatedData  // Private pointer - shared with ValidatedOptions
}
```

### Key Improvements

1. **Private Data Encapsulation**: Validated data is held in a private struct that cannot be instantiated outside the package
2. **Controlled Access**: All data access is through getter methods only
3. **Type Safety**: Impossible to bypass validation or access unvalidated data
4. **Immutable Sharing**: ValidatedOptions and CompletedOptions share data safely
5. **Compile-time Safety**: Private fields prevent accidental modification

## 3. Rationale & Context

The rationale for using this pattern (RawOptions → Validate() → Complete()) in options construction is to enforce a clear, safe, and extensible flow for handling user input and configuration in command-line tools or applications. Here’s why this pattern is beneficial:

1. Separation of Concerns

  * **RawOptions**: Collect all input (from flags, files, etc.) without assuming validity.
  * **Validate()**: Explicitly check that user input is correct and complete, catching errors early.
  * **Complete()**: Finalize and enrich validated input (e.g., loading files, resolving defaults) to produce an object safe for use.

2. Safety and Robustness

  * Code cannot accidentally use incomplete or invalid options, because only ValidatedOptions can be completed, and only CompletedOptions are used in execution.
  * Errors are caught at the correct phase, improving debuggability and user feedback.
  
3. Extensibility and Maintainability

  * New validation logic or completion steps can be added without breaking the flow.
  * The pattern scales well as more options or complexity are added.

4. Clear API Contracts

  * Each phase exposes only what’s safe for that phase, preventing misuse and making code easier to reason about.

## 4. Examples

### Good Example (Do)

```golang
opts := DefaultOptions()

// Bind CLI flags to opts (e.g., using cobra)
cmd := &cobra.Command{ /* ... */ }
BindOptions(opts, cmd)

// Validate the raw options
validated, err := opts.Validate()
if err != nil {
    // handle validation error
}

// Complete the options (e.g., load config files)
finalOpts, err := validated.Complete()
if err != nil {
    // handle completion error
}

// Now, finalOpts is ready to use with fully validated and completed data
// Access data through getter methods only
config := finalOpts.ConfigFlags()
tokenOpts := finalOpts.TokenOptions()
```

### Bad Example (Don't / Avoid)

```golang
// AVOID: Trying to access private data directly
opts := &ValidatedOptions{
    validated: &validatedData{...}, // ERROR: Cannot access private field
}

// AVOID: Bypassing validation
raw := DefaultOptions()
// directly using raw options without validation - unsafe!
UseOptions(raw) // This should never happen

// AVOID: Modifying validated data (impossible with new pattern)
validated, _ := raw.Validate()
// validated.validated.tokenOptions = nil // ERROR: Cannot access private field
```

## 5. Implementation Details

### Method Signatures

```golang
// RawOptions creation
func DefaultOptions() *Options

// Validation phase
func (o *Options) Validate() (*ValidatedOptions, error)

// Completion phase
func (vo *ValidatedOptions) Complete() (*CompletedOptions, error)

// Getter methods for ValidatedOptions
func (vo *ValidatedOptions) ConfigFlags() genericclioptions.RESTClientGetter
func (vo *ValidatedOptions) Flags() *pflag.FlagSet
func (vo *ValidatedOptions) TokenOptions() *token.Options
func (vo *ValidatedOptions) Context() string
func (vo *ValidatedOptions) AzureConfigDir() string

// Getter methods for CompletedOptions (same signatures)
func (co *CompletedOptions) ConfigFlags() genericclioptions.RESTClientGetter
func (co *CompletedOptions) Flags() *pflag.FlagSet
func (co *CompletedOptions) TokenOptions() *token.Options
func (co *CompletedOptions) Context() string
func (co *CompletedOptions) AzureConfigDir() string
```

### Key Implementation Rules

1. **Private Encapsulation**: The `validatedData` struct must be private (lowercase) and cannot be instantiated outside the package.

2. **Shared Immutable Data**: `ValidatedOptions` and `CompletedOptions` share the same `validatedData` instance via pointer, ensuring consistency without copying.

3. **Getter-Only Access**: All access to validated data must go through getter methods; no direct field access is allowed.

4. **Validation Integrity**: Once `Validate()` succeeds, the validated data cannot be modified, ensuring integrity throughout the lifecycle.

5. **Type Safety**: The compiler prevents bypassing validation phases or accessing unvalidated data.

### Benefits Achieved

- **Compile-time Safety**: Private fields prevent accidental modification or bypassing validation
- **API Clarity**: Clear progression from raw → validated → completed options
- **Memory Efficiency**: Validated data is shared, not copied, between ValidatedOptions and CompletedOptions
- **Maintainability**: Encapsulation makes it easier to modify internal data structures without breaking consumers
- **Testability**: Each phase can be tested independently with clear boundaries

## 6. Related Specifications / Further Reading

- **Domain Knowledge**: `/github/.copilot/domain_knowledge/` (when available)
- **Breadcrumb**: `/github/.copilot/breadcrumbs/2025-06-28-1921-options-pattern-improvement.md`
- **Implementation**: `/pkg/internal/converter/options_new.go`
- **Tests**: `/pkg/internal/converter/options_new_test.go`

## 7. Keywords

cli, commandline, flags, options, validation, encapsulation, type-safety, immutability