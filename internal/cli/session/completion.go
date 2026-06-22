package session

import "github.com/spf13/cobra"

// noFileCompletion suppresses Cobra's default filename completion for free-form
// flag values (e.g. templates) that have no meaningful candidate set.
func noFileCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}
