package pane

import (
	"strings"

	"github.com/MunifTanjim/tmux-ctrl/internal/cli/pane/extract"
	"github.com/MunifTanjim/tmux-ctrl/internal/tmux"
	"github.com/MunifTanjim/tmux-ctrl/internal/tui"
	"github.com/spf13/cobra"
)

func ExtractCommand() *cobra.Command {
	var (
		paneID       string
		patternNames []string
		overlay      bool
	)

	cmd := &cobra.Command{
		Use:   "extract",
		Short: "Extract tokens from a pane",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			paneID, err := resolvePaneID(paneID)
			if err != nil {
				return err
			}

			patterns, err := extract.PreparePatterns(patternNames)
			if err != nil {
				return err
			}

			content, err := tmux.CapturePane(&tmux.CapturePaneParams{TargetPane: paneID, Join: true})
			if err != nil {
				return err
			}

			cLines := strings.Split(content, "\n")
			matches := extract.Match(cLines, patterns)
			if len(matches) == 0 {
				tui.StdErrLn("No matches")
				return nil
			}

			var values []string
			if overlay {
				values, err = extract.Pick(extract.PickerHintOverlay, cLines, matches)
			} else {
				values, err = extract.Pick(extract.PickerSelect, cLines, matches)
			}
			if err != nil {
				return err
			}

			lines := make([]any, len(values))
			for i, value := range values {
				lines[i] = value
			}
			tui.StdOut(tui.Lines(lines...))
			return nil
		},
	}

	cmd.Flags().StringVarP(&paneID, "pane-id", "p", "", "pane to extract from (default: current pane)")
	cmd.Flags().StringSliceVar(&patternNames, "pattern", nil, "limit to these pattern types")
	cmd.Flags().BoolVar(&overlay, "overlay", false, "pick with an in-pane hint overlay instead of fzf")

	_ = cmd.RegisterFlagCompletionFunc("pane-id", paneCompletion(false))
	_ = cmd.RegisterFlagCompletionFunc("pattern", extract.PatternNameCompletion)

	return cmd
}
