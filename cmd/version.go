package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version, commit and build date",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			root := command.Root()
			_, err := fmt.Fprintf(command.OutOrStdout(), "%s version %s\n", root.Name(), root.Version)
			return err
		},
	}
}
