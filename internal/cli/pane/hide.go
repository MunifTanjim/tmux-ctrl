package pane

import (
	"github.com/spf13/cobra"
)

func HideCommand() *cobra.Command {
	var (
		paneID string
		tag    string
	)

	cmd := &cobra.Command{
		Use:   "hide",
		Short: "Hide a pane into the hidden session",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return hidePane(paneID, tag)
		},
	}

	cmd.Flags().StringVarP(&paneID, "pane-id", "p", "", "target pane id (default: current pane)")
	cmd.Flags().StringVar(&tag, "tag", "", "tag to add to the pane before hiding")

	_ = cmd.RegisterFlagCompletionFunc("pane-id", paneCompletion(false))

	return cmd
}
