package window

import (
	"github.com/MunifTanjim/tmux-ctrl/internal/tmux"
	"github.com/MunifTanjim/tmux-ctrl/internal/tui"
	"github.com/spf13/cobra"
)

func WidthCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "width",
		Short: "Print the current window width (columns)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			width, err := tmux.DisplayMessage("#{window_width}", &tmux.DisplayMessageParams{})
			if err != nil {
				return err
			}

			tui.StdOutLn(width)
			return nil
		},
	}
}
