package pane

import (
	"slices"
	"strings"

	"github.com/MunifTanjim/tmux-ctrl/internal/tmux"
	"github.com/MunifTanjim/tmux-ctrl/internal/tui"
	"github.com/spf13/cobra"
)

const defaultPaneListTemplate = "#{pane_id}"

func ListCommand() *cobra.Command {
	var (
		template string
		hidden   bool
		tag      string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List panes in the current window",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			lines, err := listPaneLines(template, hidden, tag)
			if err != nil {
				return err
			}

			for _, line := range lines {
				tui.StdOutLn(line)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&template, "template", defaultPaneListTemplate, "tmux format string for each pane")
	cmd.Flags().BoolVar(&hidden, "hidden", false, "list the current window's hidden panes instead of its visible ones")
	cmd.Flags().StringVar(&tag, "tag", "", "only list panes that have this tag")

	_ = cmd.RegisterFlagCompletionFunc("tag", tagFlagCompletion)
	_ = cmd.RegisterFlagCompletionFunc("template", noFileCompletion)

	return cmd
}

// listPaneLines returns the formatted pane lines for the current window: the
// hidden panes when hidden is set, otherwise the visible ones. When tag is set,
// only panes carrying that tag are kept.
func listPaneLines(template string, hidden bool, tag string) ([]string, error) {
	// When filtering, prefix a throwaway tags column so membership can be checked
	// independently of what the user's template prints.
	format := template
	if tag != "" {
		format = "#{" + paneTagsOption + "}\t" + template
	}

	lines, err := rawPaneLines(format, hidden)
	if err != nil {
		return nil, err
	}

	if tag == "" {
		return lines, nil
	}

	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		tagsCol, rest, _ := strings.Cut(line, "\t")
		if slices.Contains(parseTags(tagsCol), tag) {
			filtered = append(filtered, rest)
		}
	}
	return filtered, nil
}

// rawPaneLines lists the current window's panes (hidden or visible) with the given
// tmux format.
func rawPaneLines(format string, hidden bool) ([]string, error) {
	if !hidden {
		return tmux.ListPanes(&tmux.ListPanesParams{Format: format})
	}

	winLoc, err := currentWindowLocation()
	if err != nil {
		return nil, err
	}
	return hiddenPaneLines(winLoc, format)
}
