# Kubelogin Converter v2 Usage Examples

**Version:** 1.0  
**Created:** 2025-06-26  
**Audience:** Kubelogin developers and contributors

## Adding a New Flag

### Step 1: Add Flag Mapping
Add entry to `pkg/internal/converter/mapper/registry.go`:

```go
{
    FlagName:         "new-flag",
    ArgumentName:     "--new-flag",
    GetValue:         func(opts *token.Options) string { return opts.NewFlag },
    ApplicableLogins: []string{token.InteractiveLogin, token.DeviceCodeLogin},
},
```

### Step 2: Update Handler (if needed)
If the flag requires special handling, update the relevant handler in `pkg/internal/converter/handlers/handlers.go`:

```go
func (h *InteractiveLoginHandler) BuildExecArgs(ctx *ConversionContext, argBuilder *builder.ExecArgsBuilder) error {
    // Add new flag handling
    argBuilder.AddOptionalArgument("--new-flag", ctx.Options.NewFlag)
    return nil
}
```

### Step 3: Add Tests
Add test cases to verify the flag works correctly:

```go
func TestNewFlagHandling(t *testing.T) {
    // Test the new flag in relevant scenarios
}
```

## Adding a New Login Method

### Step 1: Define Handler
Create a new handler implementing the `LoginMethodHandler` interface:

```go
type NewLoginHandler struct {
    BaseHandler
}

func NewNewLoginHandler() *NewLoginHandler {
    return &NewLoginHandler{
        BaseHandler: BaseHandler{
            name:          "newlogin",
            requiredFlags: []string{"server-id", "client-id"},
            optionalFlags: []string{"environment"},
        },
    }
}

func (h *NewLoginHandler) Validate(ctx *ConversionContext) ValidationResult {
    return h.BaseHandler.Validate(ctx)
}

func (h *NewLoginHandler) BuildExecArgs(ctx *ConversionContext, argBuilder *builder.ExecArgsBuilder) error {
    // Build arguments specific to this login method
    return nil
}
```

### Step 2: Register Handler
Add the handler to the registry in `pkg/internal/converter/handlers/handlers.go`:

```go
func NewHandlerRegistry() map[string]LoginMethodHandler {
    return map[string]LoginMethodHandler{
        // ...existing handlers...
        token.NewLogin: NewNewLoginHandler(),
    }
}
```

### Step 3: Add Comprehensive Tests
Create test functions for validation and argument building:

```go
func TestNewLoginHandler(t *testing.T) { /* test basic functionality */ }
func TestNewLoginHandlerValidation(t *testing.T) { /* test validation */ }
func TestNewLoginHandlerBuildExecArgs(t *testing.T) { /* test argument building */ }
```

## Common Builder Patterns

### Required Arguments
```go
argBuilder.AddRequiredArgument("--server-id", serverID)
```

### Optional Arguments (only if not empty)
```go
argBuilder.AddOptionalArgument("--environment", environment)
```

### Conditional Flags
```go
argBuilder.AddFlag("--legacy", isLegacyMode)
```

### Convenience Builders
```go
// Required authentication arguments
argBuilder.AddRequiredAuthArgs(builder.RequiredAuthArgs{
    ServerID: ctx.Options.ServerID,
    ClientID: ctx.Options.ClientID,
    TenantID: ctx.Options.TenantID,
})

// Interactive-specific arguments
argBuilder.AddInteractiveArgs(builder.InteractiveArgs{
    RedirectURL: ctx.Options.RedirectURL,
    LoginHint:   ctx.Options.LoginHint,
})

// Service principal arguments
argBuilder.AddServicePrincipalArgs(builder.ServicePrincipalArgs{
    ClientSecret:            ctx.Options.ClientSecret,
    ClientCertificate:       ctx.Options.ClientCert,
    ClientCertificatePassword: ctx.Options.ClientCertPassword,
})
```

## Validation Patterns

### Basic Validation
```go
func (h *CustomHandler) Validate(ctx *ConversionContext) ValidationResult {
    // Use base validation for required fields
    if result := h.BaseHandler.Validate(ctx); !result.IsValid {
        return result
    }
    
    // Add custom validation
    if ctx.Options.CustomField == "" {
        return ValidationResult{
            IsValid: false,
            Errors:  []string{"custom-field is required for custom login"},
        }
    }
    
    return ValidationResult{IsValid: true}
}
```

### Cross-field Validation
```go
// Validate PoP token requirements
if ctx.Options.IsPoPTokenEnabled && ctx.Options.PoPTokenClaims == "" {
    return ValidationResult{
        IsValid: false,
        Errors:  []string{"--pop-claims is required when specifying --pop-enabled"},
    }
}
```

## Testing Patterns

### Handler Tests
```go
func TestCustomHandlerBuildExecArgs(t *testing.T) {
    tests := []struct {
        name     string
        options  *token.Options
        isSet    map[string]bool
        expected []string
    }{
        {
            name: "minimal custom login",
            options: &token.Options{
                ServerID: "test-server",
                ClientID: "test-client",
            },
            isSet: map[string]bool{
                "server-id": true,
                "client-id": true,
            },
            expected: []string{
                "get-token",
                "--server-id", "test-server", 
                "--client-id", "test-client",
                "--login", "custom",
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            handler := NewCustomHandler()
            ctx := &ConversionContext{
                Options: tt.options,
                IsSet:   func(flag string) bool { return tt.isSet[flag] },
            }
            
            builder := builder.NewExecArgsBuilder()
            err := handler.BuildExecArgs(ctx, builder)
            
            require.NoError(t, err)
            result := builder.MustBuild()
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Registry Tests
```go
func TestFlagMapping(t *testing.T) {
    registry := mapper.NewRegistry()
    mapping := registry.GetMapping("custom-flag")
    
    assert.NotNil(t, mapping)
    assert.Equal(t, "--custom-flag", mapping.ArgumentName)
    assert.Contains(t, mapping.ApplicableLogins, token.CustomLogin)
}
```

## Best Practices

### Handler Design
1. **Keep handlers focused** - each should handle one login method
2. **Use base validation** - extend `BaseHandler` for common requirements
3. **Validate early** - check requirements in `Validate()` method
4. **Use convenience builders** - leverage pre-built argument groups
5. **Test thoroughly** - cover all scenarios including edge cases

### Flag Registry
1. **Use descriptive names** - flag names should be self-explanatory  
2. **Specify applicable logins** - only include relevant login methods
3. **Provide value extractors** - use appropriate functions to get values
4. **Keep alphabetical** - maintain sorted order for readability

### Testing
1. **Test independently** - each component should have focused tests
2. **Use table-driven tests** - cover multiple scenarios efficiently
3. **Mock IsSet function** - control which flags are explicitly set
4. **Verify argument order** - ensure consistent output format
5. **Test error cases** - validate failure scenarios and error messages

## Debugging

### Common Issues
- **Missing arguments**: Check if flag is in registry and handler includes it
- **Wrong argument order**: Verify builder usage follows expected patterns
- **Validation failures**: Ensure required fields are provided and cross-validation passes
- **Handler not found**: Confirm handler is registered in handler registry

### Debugging Steps
1. Check registry for flag mapping
2. Verify handler implements interface correctly
3. Run specific handler tests to isolate issues
4. Use builder's `Build()` method to get detailed errors
5. Compare with working examples in test files
