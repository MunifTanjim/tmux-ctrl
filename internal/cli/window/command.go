package window

import "github.com/spf13/cobra"

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "window",
		Short: "Manage tmux windows",
	}

	cmd.AddCommand(WidthCommand())

	return cmd
}
