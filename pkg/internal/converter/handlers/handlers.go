// Package handlers provides login method strategy pattern for kubelogin converter
// This replaces the massive switch statement in Convert() with individual handlers

package handlers

import (
	"fmt"
	"github.com/Azure/kubelogin/pkg/internal/converter/builder"
	"github.com/Azure/kubelogin/pkg/internal/converter/mapper"
	"github.com/Azure/kubelogin/pkg/internal/token"
	"k8s.io/client-go/tools/clientcmd/api"
)

// ConversionContext contains all the context needed for conversion
type ConversionContext struct {
	Options           *token.Options
	AuthInfo          *api.AuthInfo
	IsLegacyProvider  bool
	FlagRegistry      *mapper.Registry
	IsSet             func(string) bool // Function to check if flag was explicitly set
}

// ValidationResult contains the result of validating conversion arguments
type ValidationResult struct {
	IsValid bool
	Errors  []string
}

// LoginMethodHandler defines the interface for handling different login methods
type LoginMethodHandler interface {
	// GetName returns the login method name this handler supports
	GetName() string
	
	// Validate checks if the conversion context is valid for this login method
	Validate(ctx *ConversionContext) ValidationResult
	
	// BuildExecArgs builds the exec arguments for this login method
	BuildExecArgs(ctx *ConversionContext, builder *builder.ExecArgsBuilder) error
	
	// GetRequiredFlags returns the flags that are required for this login method
	GetRequiredFlags() []string
	
	// GetOptionalFlags returns the flags that are optional for this login method
	GetOptionalFlags() []string
}

// BaseHandler provides common functionality for all login method handlers
type BaseHandler struct {
	name         string
	requiredFlags []string
	optionalFlags []string
}

// GetName returns the login method name
func (h *BaseHandler) GetName() string {
	return h.name
}

// GetRequiredFlags returns required flags
func (h *BaseHandler) GetRequiredFlags() []string {
	return h.requiredFlags
}

// GetOptionalFlags returns optional flags
func (h *BaseHandler) GetOptionalFlags() []string {
	return h.optionalFlags
}

// Validate performs basic validation common to all handlers
func (h *BaseHandler) Validate(ctx *ConversionContext) ValidationResult {
	var errors []string
	
	// Check required flags have values
	for _, flagName := range h.requiredFlags {
		mapping, exists := ctx.FlagRegistry.GetMapping(flagName)
		if !exists {
			continue
		}
		
		var value string
		if mapping.IsBoolean {
			// Boolean flags don't need values, just check if they should be set
			continue
		} else {
			value = mapping.GetValue(ctx.Options)
		}
		
		if value == "" {
			errors = append(errors, fmt.Sprintf("--%s is required for %s login", flagName, h.name))
		}
	}
	
	return ValidationResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}
}

// InteractiveLoginHandler handles interactive browser login
type InteractiveLoginHandler struct {
	BaseHandler
}

// NewInteractiveLoginHandler creates a new interactive login handler
func NewInteractiveLoginHandler() *InteractiveLoginHandler {
	return &InteractiveLoginHandler{
		BaseHandler: BaseHandler{
			name:         token.InteractiveLogin,
			requiredFlags: []string{"server-id", "client-id", "tenant-id"},
			optionalFlags: []string{"environment", "login-hint", "redirect-url", "pop-enabled", "pop-claims"},
		},
	}
}

// BuildExecArgs builds exec arguments for interactive login
func (h *InteractiveLoginHandler) BuildExecArgs(ctx *ConversionContext, builder *builder.ExecArgsBuilder) error {
	// Add required arguments
	mappings := ctx.FlagRegistry.GetMappingsForLogin(token.InteractiveLogin)
	
	for _, mapping := range mappings {
		if mapping.IsBoolean {
			value := mapping.GetBoolValue(ctx.Options)
			builder.AddFlag(mapping.ArgumentName, value)
		} else {
			value := mapping.GetValue(ctx.Options)
			if mapping.IsRequired || value != "" {
				builder.AddArgument(mapping.ArgumentName, value)
			}
		}
	}
	
	return nil
}

// DeviceCodeLoginHandler handles device code login
type DeviceCodeLoginHandler struct {
	BaseHandler
}

// NewDeviceCodeLoginHandler creates a new device code login handler
func NewDeviceCodeLoginHandler() *DeviceCodeLoginHandler {
	return &DeviceCodeLoginHandler{
		BaseHandler: BaseHandler{
			name:         token.DeviceCodeLogin,
			requiredFlags: []string{"server-id", "client-id", "tenant-id"},
			optionalFlags: []string{"environment", "legacy"},
		},
	}
}

// BuildExecArgs builds exec arguments for device code login
func (h *DeviceCodeLoginHandler) BuildExecArgs(ctx *ConversionContext, builder *builder.ExecArgsBuilder) error {
	mappings := ctx.FlagRegistry.GetMappingsForLogin(token.DeviceCodeLogin)
	
	for _, mapping := range mappings {
		if mapping.IsBoolean {
			value := mapping.GetBoolValue(ctx.Options)
			builder.AddFlag(mapping.ArgumentName, value)
		} else {
			value := mapping.GetValue(ctx.Options)
			if mapping.IsRequired || value != "" {
				builder.AddArgument(mapping.ArgumentName, value)
			}
		}
	}
	
	return nil
}

// ServicePrincipalLoginHandler handles service principal login
type ServicePrincipalLoginHandler struct {
	BaseHandler
}

// NewServicePrincipalLoginHandler creates a new service principal login handler
func NewServicePrincipalLoginHandler() *ServicePrincipalLoginHandler {
	return &ServicePrincipalLoginHandler{
		BaseHandler: BaseHandler{
			name:         token.ServicePrincipalLogin,
			requiredFlags: []string{"server-id", "client-id", "tenant-id"},
			optionalFlags: []string{"environment", "client-secret", "client-certificate", "client-certificate-password", "legacy", "pop-enabled", "pop-claims", "disable-environment-override"},
		},
	}
}

// BuildExecArgs builds exec arguments for service principal login
func (h *ServicePrincipalLoginHandler) BuildExecArgs(ctx *ConversionContext, builder *builder.ExecArgsBuilder) error {
	mappings := ctx.FlagRegistry.GetMappingsForLogin(token.ServicePrincipalLogin)
	
	for _, mapping := range mappings {
		if mapping.IsBoolean {
			value := mapping.GetBoolValue(ctx.Options)
			builder.AddFlag(mapping.ArgumentName, value)
		} else {
			value := mapping.GetValue(ctx.Options)
			if mapping.IsRequired || value != "" {
				builder.AddArgument(mapping.ArgumentName, value)
			}
		}
	}
	
	return nil
}

// MSILoginHandler handles managed service identity login
type MSILoginHandler struct {
	BaseHandler
}

// NewMSILoginHandler creates a new MSI login handler
func NewMSILoginHandler() *MSILoginHandler {
	return &MSILoginHandler{
		BaseHandler: BaseHandler{
			name:         token.MSILogin,
			requiredFlags: []string{"server-id"},
			optionalFlags: []string{"client-id", "identity-resource-id"},
		},
	}
}

// BuildExecArgs builds exec arguments for MSI login
func (h *MSILoginHandler) BuildExecArgs(ctx *ConversionContext, builder *builder.ExecArgsBuilder) error {
	mappings := ctx.FlagRegistry.GetMappingsForLogin(token.MSILogin)
	
	for _, mapping := range mappings {
		value := mapping.GetValue(ctx.Options)
		if mapping.IsRequired || value != "" {
			builder.AddArgument(mapping.ArgumentName, value)
		}
	}
	
	return nil
}

// Registry for all login method handlers
type HandlerRegistry struct {
	handlers map[string]LoginMethodHandler
}

// NewHandlerRegistry creates a new handler registry with all default handlers
func NewHandlerRegistry() *HandlerRegistry {
	registry := &HandlerRegistry{
		handlers: make(map[string]LoginMethodHandler),
	}
	
	// Register all handlers
	registry.Register(NewInteractiveLoginHandler())
	registry.Register(NewDeviceCodeLoginHandler())
	registry.Register(NewServicePrincipalLoginHandler())
	registry.Register(NewMSILoginHandler())
	// TODO: Add other handlers (AzureCLI, WorkloadIdentity, ROPC, AzureDeveloperCLI)
	
	return registry
}

// Register adds a handler to the registry
func (r *HandlerRegistry) Register(handler LoginMethodHandler) {
	r.handlers[handler.GetName()] = handler
}

// GetHandler returns the handler for the given login method
func (r *HandlerRegistry) GetHandler(loginMethod string) (LoginMethodHandler, bool) {
	handler, exists := r.handlers[loginMethod]
	return handler, exists
}

// GetAllHandlers returns all registered handlers
func (r *HandlerRegistry) GetAllHandlers() map[string]LoginMethodHandler {
	return r.handlers
}
