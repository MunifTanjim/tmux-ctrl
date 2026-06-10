package session

import "github.com/spf13/cobra"

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage tmux sessions",
	}

	cmd.AddCommand(NextCommand())
	cmd.AddCommand(PrevCommand())
	cmd.AddCommand(SwitchCommand())

	return cmd
}
