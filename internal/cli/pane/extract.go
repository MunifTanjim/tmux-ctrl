package pane

import (
	"strconv"
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

			capture := &tmux.CapturePaneParams{TargetPane: paneID, Join: true}
			// In copy mode the user may have scrolled up; capture the visible
			// window (shifted into history by the scroll offset), not the live bottom.
			if start, end, ok := copyModeWindow(paneID); ok {
				capture.StartLine = &start
				capture.EndLine = &end
			}

			content, err := tmux.CapturePane(capture)
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

// copyModeWindow returns the capture-pane -S/-E bounds for what a pane shows in
// copy mode; ok=false when not in copy mode or the scroll offset can't be read.
// Line 0 is the live screen's top, so scrolling up N gives lines [-N, height-1-N].
func copyModeWindow(paneID string) (start, end int, ok bool) {
	out, err := tmux.DisplayMessage(
		"#{pane_in_mode},#{scroll_position},#{pane_height}",
		&tmux.DisplayMessageParams{TargetPane: paneID},
	)
	if err != nil {
		return 0, 0, false
	}

	fields := strings.Split(strings.TrimSpace(out), ",")
	if len(fields) != 3 || fields[0] != "1" {
		return 0, 0, false
	}

	scroll, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0, 0, false
	}
	height, err := strconv.Atoi(fields[2])
	if err != nil || height <= 0 {
		return 0, 0, false
	}

	return -scroll, height - 1 - scroll, true
}
