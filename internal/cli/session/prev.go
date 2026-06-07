package session

import (
	"github.com/spf13/cobra"
)

func PrevCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "prev",
		Short: "Switch to the previous tmux session",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return switchRelative(-1)
		},
	}
}
