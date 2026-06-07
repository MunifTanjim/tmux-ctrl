package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MunifTanjim/tmux-ctrl/internal/cli/completion"
	"github.com/MunifTanjim/tmux-ctrl/internal/config"
	"github.com/MunifTanjim/tmux-ctrl/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
)

var rootCmd = &cobra.Command{
	Use:   "tmux-ctrl",
	Short: "tmux control CLI",
	Long:  `tmux-ctrl - a command-line tool for controlling tmux.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		if err := util.EnsureTool("tmux"); err != nil {
			return err
		}
		return initializeConfig(cmd)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("config file (default is $XDG_CONFIG_HOME/%s/%s)", config.ProjectName, config.ConfigFileName))
	_ = rootCmd.MarkPersistentFlagFilename("config", "yaml", "yml")

	rootCmd.InitDefaultCompletionCmd()
	if cmd, _, _ := rootCmd.Find([]string{"completion"}); cmd != nil && cmd.Name() == "completion" {
		cmd.AddCommand(completion.InstallCommand())
	}
}

func initializeConfig(cmd *cobra.Command) error {
	viper.SetEnvPrefix("TMUX_CTRL")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "*", "-", "*"))

	viper.AutomaticEnv()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(config.ConfigDir)
		viper.SetConfigName(strings.TrimSuffix(config.ConfigFileName, filepath.Ext(config.ConfigFileName)))
	}

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return err
		}
	}

	return viper.BindPFlags(cmd.Flags())
}
