# What is Kubelogin

`kubelogin` is a golang commandline application implementing azure authentication to be used by kubectl via client-go credentials plugin. It provides features that are not available in kubectl such as using `azurecli`, `spn`, `workloadidentity` login, etc.

Most of the "login modes" are the implementation of [`azidentity`](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azidentity) authentication flow.

`kubelogin` provides two main sub-commands `convert-kubeconfig` (pkg/cmd/convert.go) and `get-token` (pkg/cmd/token.go). The `get-token` sub-command gets the Entra ID access token based on the specified login mode (--login) in the kubeconfig. The `convert-kubeconfig` sub-command is a helper command to convert the kubeconfig to the specified login modes. The conversion may be lossy as some information may not be needed by the credential mode. The `convert-kubeconfig` command shares the exact command options as `get-token` command in `pkg/internal/token/options.go`. Besides using the command options, some options may be set via environment variable. The conversion code is in `pkg/internal/converter/convert.go`.

## Converter Architecture (v2.0 - June 2025)

The converter has been significantly refactored to use a modular, strategy-pattern based architecture:

### Key Components

1. **Flag Registry** (`pkg/internal/converter/mapper/registry.go`): Declarative mapping of CLI flags to arguments, eliminating repetitive if-else chains

2. **Login Method Handlers** (`pkg/internal/converter/handlers/handlers.go`): Strategy pattern implementation with 8 specialized handlers for different authentication methods:
   - InteractiveLoginHandler (browser-based auth)
   - DeviceCodeLoginHandler (device code flow)  
   - ServicePrincipalLoginHandler (SPN authentication)
   - MSILoginHandler (Managed Service Identity)
   - AzureCLILoginHandler (Azure CLI integration)
   - WorkloadIdentityLoginHandler (workload identity)
   - ROPCLoginHandler (username/password)
   - AzureDeveloperCLILoginHandler (Azure Developer CLI)

3. **Argument Builder** (`pkg/internal/converter/builder/builder.go`): Type-safe fluent interface for constructing exec arguments, replacing manual string slice construction

4. **Options Pattern** (`pkg/internal/converter/options_new.go`): Implements RawOptions → Validate() → Complete() pattern for robust configuration handling

### Benefits of New Architecture

- **Maintainability**: Adding new flags requires changes to only 1-2 files (vs 4+ previously)
- **Type Safety**: Eliminates manual string construction errors
- **Testability**: Each component is independently testable with focused responsibilities  
- **Modularity**: Login method handlers are self-contained (< 50 lines each)
- **Compatibility**: All existing tests pass - no breaking changes