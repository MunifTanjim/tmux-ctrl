package config

import (
	"github.com/MunifTanjim/tmux-ctrl/internal/config"
	"github.com/MunifTanjim/tmux-ctrl/internal/tui"
	"github.com/spf13/cobra"
)

func DirCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "dir",
		Short: "Print the config directory path",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			tui.StdOut(config.ConfigDir)
			return nil
		},
	}
}
