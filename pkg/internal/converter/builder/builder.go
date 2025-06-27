// Package builder provides type-safe exec argument construction for kubelogin converter
// This replaces manual string slice construction with a fluent interface

package builder

import (
	"fmt"
)

// ExecArgsBuilder provides a fluent interface for building exec arguments
type ExecArgsBuilder struct {
	args []string
}

// NewExecArgsBuilder creates a new exec args builder with the base command
func NewExecArgsBuilder() *ExecArgsBuilder {
	return &ExecArgsBuilder{
		args: []string{"get-token"},
	}
}

// AddArgument adds a flag-value pair to the arguments
func (b *ExecArgsBuilder) AddArgument(flag, value string) *ExecArgsBuilder {
	if value == "" {
		return b // Skip empty values
	}
	b.args = append(b.args, flag, value)
	return b
}

// AddRequiredArgument adds a required flag-value pair, panicking if value is empty
func (b *ExecArgsBuilder) AddRequiredArgument(flag, value string) *ExecArgsBuilder {
	if value == "" {
		panic(fmt.Sprintf("required argument %s cannot be empty", flag))
	}
	b.args = append(b.args, flag, value)
	return b
}

// AddFlag adds a boolean flag (flag without value) if condition is true
func (b *ExecArgsBuilder) AddFlag(flag string, condition bool) *ExecArgsBuilder {
	if condition {
		b.args = append(b.args, flag)
	}
	return b
}

// AddOptionalArgument adds a flag-value pair only if value is not empty
func (b *ExecArgsBuilder) AddOptionalArgument(flag, value string) *ExecArgsBuilder {
	if value != "" {
		b.args = append(b.args, flag, value)
	}
	return b
}

// Build returns the final arguments slice
func (b *ExecArgsBuilder) Build() []string {
	// Return a copy to prevent external modification
	result := make([]string, len(b.args))
	copy(result, b.args)
	return result
}

// Length returns the current number of arguments
func (b *ExecArgsBuilder) Length() int {
	return len(b.args)
}

// Reset clears all arguments except the base command
func (b *ExecArgsBuilder) Reset() *ExecArgsBuilder {
	b.args = []string{"get-token"}
	return b
}

// Clone creates a copy of the current builder
func (b *ExecArgsBuilder) Clone() *ExecArgsBuilder {
	newBuilder := &ExecArgsBuilder{
		args: make([]string, len(b.args)),
	}
	copy(newBuilder.args, b.args)
	return newBuilder
}

// String returns a string representation of the current arguments
func (b *ExecArgsBuilder) String() string {
	return fmt.Sprintf("ExecArgs[%d]: %v", len(b.args), b.args)
}
