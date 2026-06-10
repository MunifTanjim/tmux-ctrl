package pane

import (
	"slices"

	"github.com/MunifTanjim/tmux-ctrl/internal/tmux"
	"github.com/spf13/cobra"
)

// completionFunc is the shape Cobra expects for flag and positional-argument
// completion.
type completionFunc = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)

// noFileCompletion suppresses Cobra's default filename completion for flags whose
// values are free-form (sizes, templates) and have no meaningful candidate set.
func noFileCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}

// paneCompletion returns a completion func offering pane ids, each annotated with
// its location and current command. When all is set it spans every session and
// window (including the hidden holding session); otherwise it is scoped to the
// current window. tmux query failures degrade to no completion.
func paneCompletion(all bool) completionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// "id\tdescription" is the form Cobra renders as a described candidate.
		lines, err := tmux.ListPanes(&tmux.ListPanesParams{
			All:    all,
			Format: "#{pane_id}\t#{session_name}:#{window_index}.#{pane_index} #{pane_current_command}",
		})
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return lines, cobra.ShellCompDirectiveNoFileComp
	}
}

// directionDescriptions annotates each placement direction in the completion menu.
var directionDescriptions = map[string]string{
	"bottom":       "split below",
	"top":          "split above",
	"right":        "split to the right",
	"left":         "split to the left",
	"top-left":     "dock to the top-left corner",
	"top-right":    "dock to the top-right corner",
	"bottom-left":  "dock to the bottom-left corner",
	"bottom-right": "dock to the bottom-right corner",
}

// directionCompletion offers the placement directions with descriptions.
func directionCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	out := make([]string, 0, len(paneDirections))
	for _, dir := range paneDirections {
		out = append(out, dir+"\t"+directionDescriptions[dir])
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}

// moveDirectionCompletion offers the shared placement directions plus swap, which
// is exclusive to `pane move`.
func moveDirectionCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	out, _ := directionCompletion(cmd, args, toComplete)
	out = append(out, swapDirection+"\t"+"swap with another pane")
	return out, cobra.ShellCompDirectiveNoFileComp
}

// collectTags returns the distinct tags in use across every pane on the server,
// in first-seen order. tmux query failures yield no tags.
func collectTags() []string {
	lines, err := tmux.ListPanes(&tmux.ListPanesParams{
		All:    true,
		Format: "#{" + paneTagsOption + "}",
	})
	if err != nil {
		return nil
	}

	var tags []string
	for _, line := range lines {
		for _, tag := range parseTags(line) {
			if !slices.Contains(tags, tag) {
				tags = append(tags, tag)
			}
		}
	}
	return tags
}

// without returns tags omitting any value present in exclude.
func without(tags, exclude []string) []string {
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		if !slices.Contains(exclude, tag) {
			out = append(out, tag)
		}
	}
	return out
}

// tagFlagCompletion offers known tags for a --tag flag value.
func tagFlagCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return collectTags(), cobra.ShellCompDirectiveNoFileComp
}

// tagRemoveArgCompletion offers the target pane's current tags for `pane tag
// remove`, dropping ones already listed on the command line.
func tagRemoveArgCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	paneID, err := resolvePaneID(cmd.Flag("pane-id").Value.String())
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	tags, err := getPaneTags(paneID)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return without(tags, args), cobra.ShellCompDirectiveNoFileComp
}
