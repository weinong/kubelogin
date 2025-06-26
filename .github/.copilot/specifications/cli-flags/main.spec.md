# Specification: Validated and Completed Secrets Pattern

**Version:** 1.0

**Last Updated:** 2025-06-26

**Owner:** weinong

## 1. Purpose & Scope

This document defines a structured process for handling command-line options by separating their construction into three clear phases.

1. **RawOptions**: Hold all raw (unvalidated) input values (e.g., from CLI flags).
2. **Validate()**: Validate the raw options, returning a **ValidatedOptions** object if checks pass.
3. **Complete()**: Complete the validated options to produce an **Options** object, loading any additional configuration needed (e.g., from files).

Each phase enforces order:
* You must call Validate() before Complete().
* Only fully constructed Options are ready for use.

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

// Now, finalOpts is ready to use
```

### Bad Example (Don't / Avoid)

```
// Code snippet demonstrating what to avoid
```

## 5. Related Specifications / Further Reading

## 6. Keywords

cli, commandline, flags, options