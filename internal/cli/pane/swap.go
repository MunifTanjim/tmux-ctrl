package pane

import (
	"github.com/MunifTanjim/tmux-ctrl/internal/tmux"
	"github.com/spf13/cobra"
)

func SwapCommand() *cobra.Command {
	var (
		paneID string
		target string
	)

	cmd := &cobra.Command{
		Use:   "swap",
		Short: "Swap a pane with another pane in its window",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			srcPaneID, err := resolvePaneID(paneID)
			if err != nil {
				return err
			}

			if err := ensureWindowUnzoomed(srcPaneID); err != nil {
				return err
			}

			if target == "" {
				target, err = pickTargetPane(paneSwapTargetOption)
				if err != nil {
					return err
				}
				if target == "" {
					return nil
				}
			}

			if target == srcPaneID {
				return nil
			}

			return tmux.SwapPane(&tmux.SwapPaneParams{
				SrcPane:  srcPaneID,
				DstPane:  target,
				Detached: true,
			})
		},
	}

	cmd.Flags().StringVarP(&paneID, "pane-id", "p", "", "pane to swap (default: current pane)")
	cmd.Flags().StringVar(&target, "target", "", "target pane id")

	_ = cmd.RegisterFlagCompletionFunc("pane-id", paneCompletion(false))
	_ = cmd.RegisterFlagCompletionFunc("target", paneCompletion(false))

	return cmd
}
