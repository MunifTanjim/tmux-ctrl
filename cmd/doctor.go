package cmd

import (
	"github.com/MunifTanjim/tmux-ctrl/internal/cli/doctor"
)

func init() {
	rootCmd.AddCommand(doctor.Command())
}
