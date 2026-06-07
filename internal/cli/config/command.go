package config

import "github.com/spf13/cobra"

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management",
		Long:  `Commands for managing tmux-ctrl configuration.`,
	}

	cmd.AddCommand(DirCommand())
	cmd.AddCommand(GetCommand())
	cmd.AddCommand(SetCommand())

	return cmd
}
