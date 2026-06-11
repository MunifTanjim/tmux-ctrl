package pane

import (
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pane",
		Short: "Manage tmux panes",
	}

	cmd.AddCommand(HideCommand())
	cmd.AddCommand(ShowCommand())
	cmd.AddCommand(MoveCommand())
	cmd.AddCommand(SplitCommand())
	cmd.AddCommand(ExtractCommand())
	cmd.AddCommand(StateCommand())
	cmd.AddCommand(ListCommand())
	cmd.AddCommand(TagCommand())

	return cmd
}
