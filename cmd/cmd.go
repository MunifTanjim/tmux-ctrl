package cmd

import (
	"github.com/MunifTanjim/tmux-ctrl/internal/cli/config"
	"github.com/MunifTanjim/tmux-ctrl/internal/cli/doctor"
	"github.com/MunifTanjim/tmux-ctrl/internal/cli/pane"
	"github.com/MunifTanjim/tmux-ctrl/internal/cli/session"
	"github.com/MunifTanjim/tmux-ctrl/internal/cli/window"
)

func init() {
	rootCmd.AddCommand(
		config.Command(),
		doctor.Command(),
		pane.Command(),
		session.Command(),
		window.Command(),
	)
}
