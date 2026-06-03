package config

import (
	"sort"

	"github.com/MunifTanjim/tmux-ctrl/internal/config"
	"github.com/MunifTanjim/tmux-ctrl/internal/tui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func GetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get [key]",
		Short: "Get configuration value(s)",
		Long:  `Get a specific configuration value by key, or list all values if no key is provided.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				keys := viper.AllKeys()
				sort.Strings(keys)
				for _, key := range keys {
					if key == "config" || key == "help" {
						continue
					}
					tui.StdOutF("%s=%v\n", key, viper.Get(key))
				}
				return nil
			}

			value := viper.Get(args[0])
			if value == nil {
				return nil
			}
			tui.StdOut(value)
			return nil
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return config.KnownKeys, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}
}
