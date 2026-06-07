package pane

import (
	"github.com/MunifTanjim/tmux-ctrl/internal/tui"
	"github.com/spf13/cobra"
)

func StateCommand() *cobra.Command {
	var paneID string

	cmd := &cobra.Command{
		Use:   "state",
		Short: "Print whether a pane is hidden or visible",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			hidden, err := isPaneHidden(paneID)
			if err != nil {
				return err
			}

			if hidden {
				tui.StdOutLn("hidden")
			} else {
				tui.StdOutLn("visible")
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&paneID, "pane-id", "p", "", "target pane id")
	cmd.MarkFlagRequired("pane-id")

	_ = cmd.RegisterFlagCompletionFunc("pane-id", paneCompletion(true))

	return cmd
}
