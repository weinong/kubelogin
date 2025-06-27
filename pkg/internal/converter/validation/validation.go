// Package validation provides centralized validation for kubelogin converter
// This replaces scattered validation logic with a schema-based approach

package validation

import (
	"fmt"
	"strings"
)

// ValidationRule represents a single validation rule
type ValidationRule interface {
	// Validate checks if the rule passes and returns error message if not
	Validate(ctx *ValidationContext) *ValidationError
	
	// GetDescription returns a human-readable description of the rule
	GetDescription() string
}

// ValidationContext contains all context needed for validation
type ValidationContext struct {
	LoginMethod string
	Values      map[string]string  // flag name -> value
	Flags       map[string]bool    // flag name -> is set
	Options     interface{}        // token.Options for complex validations
}

// ValidationError represents a validation failure
type ValidationError struct {
	Rule        string
	Message     string
	FieldName   string
	LoginMethod string
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.LoginMethod, e.FieldName, e.Message)
}

// ValidationResult contains all validation results
type ValidationResult struct {
	IsValid bool
	Errors  []*ValidationError
}

// AddError adds a validation error
func (r *ValidationResult) AddError(err *ValidationError) {
	r.Errors = append(r.Errors, err)
	r.IsValid = false
}

// GetErrorMessages returns all error messages as strings
func (r *ValidationResult) GetErrorMessages() []string {
	var messages []string
	for _, err := range r.Errors {
		messages = append(messages, err.Error())
	}
	return messages
}

// RequiredFieldRule validates that a required field has a value
type RequiredFieldRule struct {
	FieldName   string
	DisplayName string
}

// NewRequiredFieldRule creates a new required field rule
func NewRequiredFieldRule(fieldName, displayName string) *RequiredFieldRule {
	return &RequiredFieldRule{
		FieldName:   fieldName,
		DisplayName: displayName,
	}
}

// Validate checks if the required field has a value
func (r *RequiredFieldRule) Validate(ctx *ValidationContext) *ValidationError {
	value, exists := ctx.Values[r.FieldName]
	if !exists || strings.TrimSpace(value) == "" {
		return &ValidationError{
			Rule:        "required_field",
			Message:     fmt.Sprintf("--%s is required", r.FieldName),
			FieldName:   r.FieldName,
			LoginMethod: ctx.LoginMethod,
		}
	}
	return nil
}

// GetDescription returns the rule description
func (r *RequiredFieldRule) GetDescription() string {
	return fmt.Sprintf("%s is required", r.DisplayName)
}

// MutuallyExclusiveRule validates that only one of a set of fields is set
type MutuallyExclusiveRule struct {
	FieldNames []string
	GroupName  string
}

// NewMutuallyExclusiveRule creates a new mutually exclusive rule
func NewMutuallyExclusiveRule(groupName string, fieldNames ...string) *MutuallyExclusiveRule {
	return &MutuallyExclusiveRule{
		FieldNames: fieldNames,
		GroupName:  groupName,
	}
}

// Validate checks that only one field in the group is set
func (r *MutuallyExclusiveRule) Validate(ctx *ValidationContext) *ValidationError {
	setFields := []string{}
	
	for _, fieldName := range r.FieldNames {
		if value, exists := ctx.Values[fieldName]; exists && strings.TrimSpace(value) != "" {
			setFields = append(setFields, fieldName)
		}
	}
	
	if len(setFields) > 1 {
		return &ValidationError{
			Rule:        "mutually_exclusive",
			Message:     fmt.Sprintf("only one of [%s] can be specified", strings.Join(setFields, ", ")),
			FieldName:   strings.Join(setFields, ","),
			LoginMethod: ctx.LoginMethod,
		}
	}
	
	return nil
}

// GetDescription returns the rule description
func (r *MutuallyExclusiveRule) GetDescription() string {
	return fmt.Sprintf("only one of [%s] can be specified", strings.Join(r.FieldNames, ", "))
}

// ConditionalRequiredRule validates that if one field is set, another is also required
type ConditionalRequiredRule struct {
	TriggerField  string
	RequiredField string
	Message       string
}

// NewConditionalRequiredRule creates a new conditional required rule
func NewConditionalRequiredRule(triggerField, requiredField, message string) *ConditionalRequiredRule {
	return &ConditionalRequiredRule{
		TriggerField:  triggerField,
		RequiredField: requiredField,
		Message:       message,
	}
}

// Validate checks the conditional requirement
func (r *ConditionalRequiredRule) Validate(ctx *ValidationContext) *ValidationError {
	triggerValue, triggerExists := ctx.Values[r.TriggerField]
	requiredValue, requiredExists := ctx.Values[r.RequiredField]
	
	// If trigger field is set and has value, required field must also be set
	if triggerExists && strings.TrimSpace(triggerValue) != "" {
		if !requiredExists || strings.TrimSpace(requiredValue) == "" {
			return &ValidationError{
				Rule:        "conditional_required",
				Message:     r.Message,
				FieldName:   r.RequiredField,
				LoginMethod: ctx.LoginMethod,
			}
		}
	}
	
	return nil
}

// GetDescription returns the rule description
func (r *ConditionalRequiredRule) GetDescription() string {
	return r.Message
}

// LoginMethodSchema defines validation rules for a specific login method
type LoginMethodSchema struct {
	LoginMethod string
	Rules       []ValidationRule
}

// NewLoginMethodSchema creates a new login method schema
func NewLoginMethodSchema(loginMethod string) *LoginMethodSchema {
	return &LoginMethodSchema{
		LoginMethod: loginMethod,
		Rules:       []ValidationRule{},
	}
}

// AddRule adds a validation rule to the schema
func (s *LoginMethodSchema) AddRule(rule ValidationRule) *LoginMethodSchema {
	s.Rules = append(s.Rules, rule)
	return s
}

// Validate validates the context against all rules in the schema
func (s *LoginMethodSchema) Validate(ctx *ValidationContext) *ValidationResult {
	result := &ValidationResult{IsValid: true}
	
	for _, rule := range s.Rules {
		if err := rule.Validate(ctx); err != nil {
			result.AddError(err)
		}
	}
	
	return result
}

// SchemaRegistry holds validation schemas for all login methods
type SchemaRegistry struct {
	schemas map[string]*LoginMethodSchema
}

// NewSchemaRegistry creates a new schema registry with default schemas
func NewSchemaRegistry() *SchemaRegistry {
	registry := &SchemaRegistry{
		schemas: make(map[string]*LoginMethodSchema),
	}
	
	// Register default schemas
	registry.registerDefaultSchemas()
	
	return registry
}

// registerDefaultSchemas registers validation schemas for all login methods
func (r *SchemaRegistry) registerDefaultSchemas() {
	// Interactive Login Schema
	interactive := NewLoginMethodSchema("interactive").
		AddRule(NewRequiredFieldRule("server-id", "Server ID")).
		AddRule(NewRequiredFieldRule("client-id", "Client ID")).
		AddRule(NewRequiredFieldRule("tenant-id", "Tenant ID")).
		AddRule(NewConditionalRequiredRule("pop-enabled", "pop-claims", "--pop-claims is required when specifying --pop-enabled"))
	
	// Device Code Login Schema
	devicecode := NewLoginMethodSchema("devicecode").
		AddRule(NewRequiredFieldRule("server-id", "Server ID")).
		AddRule(NewRequiredFieldRule("client-id", "Client ID")).
		AddRule(NewRequiredFieldRule("tenant-id", "Tenant ID"))
	
	// Service Principal Login Schema
	spn := NewLoginMethodSchema("spn").
		AddRule(NewRequiredFieldRule("server-id", "Server ID")).
		AddRule(NewRequiredFieldRule("client-id", "Client ID")).
		AddRule(NewRequiredFieldRule("tenant-id", "Tenant ID")).
		AddRule(NewMutuallyExclusiveRule("authentication", "client-secret", "client-certificate")).
		AddRule(NewConditionalRequiredRule("pop-enabled", "pop-claims", "--pop-claims is required when specifying --pop-enabled"))
	
	// MSI Login Schema
	msi := NewLoginMethodSchema("msi").
		AddRule(NewRequiredFieldRule("server-id", "Server ID")).
		AddRule(NewMutuallyExclusiveRule("identity", "client-id", "identity-resource-id"))
	
	// Register schemas
	r.RegisterSchema(interactive)
	r.RegisterSchema(devicecode)
	r.RegisterSchema(spn)
	r.RegisterSchema(msi)
}

// RegisterSchema registers a schema for a login method
func (r *SchemaRegistry) RegisterSchema(schema *LoginMethodSchema) {
	r.schemas[schema.LoginMethod] = schema
}

// GetSchema returns the schema for a login method
func (r *SchemaRegistry) GetSchema(loginMethod string) (*LoginMethodSchema, bool) {
	schema, exists := r.schemas[loginMethod]
	return schema, exists
}

// ValidateLoginMethod validates a login method using its registered schema
func (r *SchemaRegistry) ValidateLoginMethod(loginMethod string, ctx *ValidationContext) *ValidationResult {
	schema, exists := r.GetSchema(loginMethod)
	if !exists {
		result := &ValidationResult{IsValid: false}
		result.AddError(&ValidationError{
			Rule:        "unknown_login_method",
			Message:     fmt.Sprintf("unknown login method: %s", loginMethod),
			FieldName:   "login",
			LoginMethod: loginMethod,
		})
		return result
	}
	
	ctx.LoginMethod = loginMethod
	return schema.Validate(ctx)
}
