package pane

import (
	"errors"
	"fmt"
	"slices"
	"strings"

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

			selected, err := tui.NewFuzzySearch(tui.FuzzySearchConfig[hiddenPane]{
				Header:        "Select pane to show",
				Items:         panes,
				AutoSelectOne: true,
				Preview: func(pane hiddenPane) string {
					if len(pane.Tags) > 0 {
						return fmt.Sprintf("%s\t%s\t%s", strings.Join(pane.Tags, ","), pane.Command, pane.Path)
					}
					return fmt.Sprintf("%s\t%s", pane.Command, pane.Path)
				},
			}).Run()
			if err != nil {
				if errors.Is(err, tui.ErrCancelled) || errors.Is(err, tui.ErrNoMatch) {
					return nil
				}
				return err
			}

			if len(selected) == 0 {
				return nil
			}

			if err := ensureWindowUnzoomed(""); err != nil {
				return err
			}

			return showPane(winLoc, selected[0].Ref, direction, size)
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
