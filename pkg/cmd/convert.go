package cmd

import (
	"github.com/Azure/kubelogin/pkg/internal/converter"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

// newConvertCmd provides a cobra command for convert sub command
func newConvertCmd() *cobra.Command {
	rawOpts := converter.NewRawOptions()

	cmd := &cobra.Command{
		Use:          "convert-kubeconfig",
		Short:        "convert kubeconfig to use exec auth module",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			rawOpts.UpdateFromEnv()

			// Validate raw options
			validatedOpts, err := rawOpts.Validate()
			if err != nil {
				return err
			}

			// Complete validated options
			completedOpts, err := validatedOpts.Complete()
			if err != nil {
				return err
			}

			pathOptions := clientcmd.NewDefaultPathOptions()
			pathOptions.LoadingRules.ExplicitPath, _ = completedOpts.GetFlags().GetString("kubeconfig")

			if err := converter.Convert(completedOpts, pathOptions); err != nil {
				return err
			}
			return nil
		},
		ValidArgsFunction: cobra.NoFileCompletions,
	}

	rawOpts.AddFlags(cmd.Flags())
	rawOpts.AddCompletions(cmd)

	return cmd
}
