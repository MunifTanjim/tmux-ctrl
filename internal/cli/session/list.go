package session

import (
	"strings"

	"github.com/MunifTanjim/tmux-ctrl/internal/config"
	"github.com/MunifTanjim/tmux-ctrl/internal/tmux"
	"github.com/MunifTanjim/tmux-ctrl/internal/tui"
	"github.com/spf13/cobra"
)

const defaultSessionListTemplate = "#{session_name}"

func ListCommand() *cobra.Command {
	var (
		template string
		hidden   bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List sessions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			lines, err := listSessionLines(template, hidden)
			if err != nil {
				return err
			}

			for _, line := range lines {
				tui.StdOutLn(line)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&template, "template", defaultSessionListTemplate, "tmux format string for each session")
	cmd.Flags().BoolVar(&hidden, "hidden", false, "include the hidden session in the output")

	_ = cmd.RegisterFlagCompletionFunc("template", noFileCompletion)

	return cmd
}

// listSessionLines returns the formatted session lines. Unless includeHidden is
// set, the hidden session is dropped.
func listSessionLines(template string, includeHidden bool) ([]string, error) {
	if includeHidden {
		return tmux.ListSessions(&tmux.ListSessionsParams{Format: template})
	}

	// Prefix a throwaway name column so the hidden session can be detected
	// independently of what the user's template prints.
	lines, err := tmux.ListSessions(&tmux.ListSessionsParams{Format: "#{session_name}\t" + template})
	if err != nil {
		return nil, err
	}

	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		name, rest, _ := strings.Cut(line, "\t")
		if name == config.HiddenSessionName {
			continue
		}
		filtered = append(filtered, rest)
	}
	return filtered, nil
}
