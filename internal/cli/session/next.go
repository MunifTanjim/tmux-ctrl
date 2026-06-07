package session

import (
	"github.com/spf13/cobra"
)

func NextCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "next",
		Short: "Switch to the next tmux session",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return switchRelative(1)
		},
	}
}
