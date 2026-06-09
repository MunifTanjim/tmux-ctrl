package pane

import (
	"github.com/spf13/cobra"
)

func MoveCommand() *cobra.Command {
	var (
		paneID    string
		direction string
		size      string
		target    string
		edge      bool
	)

	cmd := &cobra.Command{
		Use:   "move",
		Short: "Reposition a pane within its window",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			paneID, err := resolvePaneID(paneID)
			if err != nil {
				return err
			}

			if err := ensureWindowUnzoomed(paneID); err != nil {
				return err
			}

			// Corners are inherently an edge placement; assume --edge for them.
			if edge || isCornerDirection(direction) {
				return movePaneToEdge(paneID, direction, size)
			}

			return movePaneRelative(paneID, target, direction, size)
		},
	}

	cmd.Flags().StringVarP(&paneID, "pane-id", "p", "", "pane to move (default: current pane)")
	cmd.Flags().StringVarP(&direction, "direction", "d", "bottom", "placement direction: "+directionsHint)
	cmd.Flags().StringVar(&size, "size", "", "size of the resulting split (lines, columns, percentage, or 0<fraction<1 as percentage)")
	cmd.Flags().StringVar(&target, "target", "", "target pane id (skips the picker; ignored with --edge)")
	cmd.Flags().BoolVar(&edge, "edge", false, "move to the window edge in --direction instead of picking a target pane")

	cmd.MarkFlagsMutuallyExclusive("edge", "target")

	_ = cmd.RegisterFlagCompletionFunc("pane-id", paneCompletion(false))
	_ = cmd.RegisterFlagCompletionFunc("target", paneCompletion(false))
	_ = cmd.RegisterFlagCompletionFunc("direction", directionCompletion)
	_ = cmd.RegisterFlagCompletionFunc("size", noFileCompletion)

	return cmd
}
