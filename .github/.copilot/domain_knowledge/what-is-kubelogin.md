# What is Kubelogin

`kubelogin` is a golang commandline application implementing azure authentication to be used by kubectl via client-go credentials plugin. It provides features that are not available in kubectl such as using `azurecli`, `spn`, `workloadidentity` login, etc.

Most of the "login modes" are the implementation of [`azidentity`](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azidentity) authentication flow.

`kubelogin` provides two main sub-commands `convert-kubeconfig` (pkg/cmd/convert.go) and `get-token` (pkg/cmd/token.go). The `get-token` sub-command gets the Entra ID access token based on the specified login mode (--login) in the kubeconfig. The `convert-kubeconfig` sub-command is a helper command to convert the kubeconfig to the specified login modes. The conversion may be lossy as some information may not be needed by the credential mode. The `convert-kubeconfig` command shares the exact command options as `get-token` command in `pkg/internal/token/options.go`. Besides using the command options, some options may be set via environment variable. The conversion code is in `pkg/internal/converter/convert.go`.