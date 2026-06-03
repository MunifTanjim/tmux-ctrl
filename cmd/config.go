package cmd

import (
	"github.com/MunifTanjim/tmux-ctrl/internal/cli/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
	Long:  `Commands for managing tmux-ctrl configuration.`,
}

func init() {
	configCmd.AddCommand(config.DirCommand())
	configCmd.AddCommand(config.GetCommand())
	configCmd.AddCommand(config.SetCommand())

	rootCmd.AddCommand(configCmd)
}
