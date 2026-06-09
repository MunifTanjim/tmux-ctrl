package pane

import (
	"slices"

	"github.com/MunifTanjim/tmux-ctrl/internal/tui"
	"github.com/spf13/cobra"
)

func ShowCommand() *cobra.Command {
	var (
		tag       string
		direction string
		size      string
	)

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Restore a hidden pane into the current window",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			winLoc, err := currentWindowLocation()
			if err != nil {
				return err
			}

			panes, err := listHiddenPanes(winLoc)
			if err != nil {
				return err
			}

			if tag != "" {
				filtered := panes[:0]
				for _, pane := range panes {
					if slices.Contains(pane.Tags, tag) {
						filtered = append(filtered, pane)
					}
				}
				panes = filtered
			}

			if len(panes) == 0 {
				tui.StdErrLn("No hidden panes")
				return nil
			}

			selectedRef, err := selectHiddenPane(panes)
			if err != nil {
				return err
			}
			if selectedRef == "" {
				return nil
			}

			if err := ensureWindowUnzoomed(""); err != nil {
				return err
			}

			return showPane(winLoc, selectedRef, direction, size)
		},
	}

	cmd.Flags().StringVar(&tag, "tag", "", "only show panes with this tag")
	cmd.Flags().StringVarP(&direction, "direction", "d", "bottom", "split direction: "+directionsHint)
	cmd.Flags().StringVar(&size, "size", "", "size of the restored pane (lines, columns, percentage, or 0<fraction<1 as percentage)")

	_ = cmd.RegisterFlagCompletionFunc("tag", tagFlagCompletion)
	_ = cmd.RegisterFlagCompletionFunc("direction", directionCompletion)
	_ = cmd.RegisterFlagCompletionFunc("size", noFileCompletion)

	return cmd
}
