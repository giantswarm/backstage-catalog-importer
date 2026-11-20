package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	t.Cleanup(func() {
		rootCmd.SetArgs(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		require.NoError(t, rootCmd.Flags().Set("version", "false"))
	})
	require.NotEmpty(t, rootCmd.Version)

	for _, argument := range []string{"version", "--version"} {
		t.Run(argument, func(t *testing.T) {
			var output, errorOutput bytes.Buffer
			rootCmd.SetArgs([]string{argument})
			rootCmd.SetOut(&output)
			rootCmd.SetErr(&errorOutput)

			require.NoError(t, rootCmd.ExecuteContext(t.Context()))
			require.Equal(t, "backstage-catalog-importer version "+rootCmd.Version+"\n", output.String())
			require.Empty(t, errorOutput.String())
		})
	}
}

func TestVersionRejectsArguments(t *testing.T) {
	command := newVersionCommand()
	command.SetArgs([]string{"unexpected"})
	command.SetOut(&bytes.Buffer{})
	command.SetErr(&bytes.Buffer{})
	require.ErrorContains(t, command.ExecuteContext(t.Context()), "unknown command")
}
