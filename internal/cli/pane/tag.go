package pane

import (
	"fmt"
	"slices"
	"strings"

	"github.com/MunifTanjim/tmux-ctrl/internal/tmux"
	"github.com/MunifTanjim/tmux-ctrl/internal/tui"
	"github.com/spf13/cobra"
)

func TagCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag",
		Short: "Manage pane tags",
	}

	cmd.AddCommand(tagListCommand())
	cmd.AddCommand(tagAddCommand())
	cmd.AddCommand(tagRemoveCommand())

	return cmd
}

func tagListCommand() *cobra.Command {
	var paneID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the pane's tags, one per line",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			paneID, err := resolvePaneID(paneID)
			if err != nil {
				return err
			}

			tags, err := getPaneTags(paneID)
			if err != nil {
				return err
			}

			for _, tag := range tags {
				tui.StdOutLn(tag)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&paneID, "pane-id", "p", "", "target pane id (default: current pane)")

	_ = cmd.RegisterFlagCompletionFunc("pane-id", paneCompletion(true))

	return cmd
}

func tagAddCommand() *cobra.Command {
	var paneID string

	cmd := &cobra.Command{
		Use:   "add <tag>...",
		Short: "Add one or more tags to the pane",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			paneID, err := resolvePaneID(paneID)
			if err != nil {
				return err
			}

			return addPaneTags(paneID, args)
		},
	}

	cmd.Flags().StringVarP(&paneID, "pane-id", "p", "", "target pane id (default: current pane)")

	_ = cmd.RegisterFlagCompletionFunc("pane-id", paneCompletion(true))

	return cmd
}

func tagRemoveCommand() *cobra.Command {
	var paneID string

	cmd := &cobra.Command{
		Use:   "remove <tag>...",
		Short: "Remove one or more tags from the pane",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			paneID, err := resolvePaneID(paneID)
			if err != nil {
				return err
			}

			return removePaneTags(paneID, args)
		},
		ValidArgsFunction: tagRemoveArgCompletion,
	}

	cmd.Flags().StringVarP(&paneID, "pane-id", "p", "", "target pane id (default: current pane)")

	_ = cmd.RegisterFlagCompletionFunc("pane-id", paneCompletion(true))

	return cmd
}

// getPaneTags reads paneID's comma-separated tags, trimmed, in stored order.
func getPaneTags(paneID string) ([]string, error) {
	value, err := tmux.DisplayMessage("#{"+paneTagsOption+"}", &tmux.DisplayMessageParams{TargetPane: paneID})
	if err != nil {
		return nil, err
	}
	return parseTags(value), nil
}

// parseTags splits a stored comma-separated tags value into trimmed, non-empty
// tags, preserving order.
func parseTags(value string) []string {
	var tags []string
	for _, tag := range strings.Split(value, ",") {
		if tag = strings.TrimSpace(tag); tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

// addPaneTags validates rawTags and adds the new ones to paneID's tags, deduped
// and preserving order.
func addPaneTags(paneID string, rawTags []string) error {
	tags, err := normalizeTags(rawTags)
	if err != nil {
		return err
	}

	existing, err := getPaneTags(paneID)
	if err != nil {
		return err
	}

	for _, tag := range tags {
		if !slices.Contains(existing, tag) {
			existing = append(existing, tag)
		}
	}

	return setPaneTags(paneID, existing)
}

// removePaneTags validates rawTags and removes them from paneID's tags,
// preserving the order of the remaining tags.
func removePaneTags(paneID string, rawTags []string) error {
	tags, err := normalizeTags(rawTags)
	if err != nil {
		return err
	}

	existing, err := getPaneTags(paneID)
	if err != nil {
		return err
	}

	remaining := existing[:0]
	for _, tag := range existing {
		if !slices.Contains(tags, tag) {
			remaining = append(remaining, tag)
		}
	}

	return setPaneTags(paneID, remaining)
}

// setPaneTags writes tags back as a comma-separated list, unsetting the option
// entirely when there are none left.
func setPaneTags(paneID string, tags []string) error {
	if len(tags) == 0 {
		return tmux.UnsetPaneOption(paneTagsOption, &tmux.SetPaneOptionParams{TargetPane: paneID})
	}
	return tmux.SetPaneOption(paneTagsOption, strings.Join(tags, ","), &tmux.SetPaneOptionParams{TargetPane: paneID})
}

// normalizeTags trims each tag and rejects empty ones or ones containing a comma
// (the list delimiter).
func normalizeTags(args []string) ([]string, error) {
	tags := make([]string, 0, len(args))
	for _, arg := range args {
		tag := strings.TrimSpace(arg)
		if tag == "" {
			return nil, fmt.Errorf("tag must not be empty")
		}
		if strings.Contains(tag, ",") {
			return nil, fmt.Errorf("tag must not contain a comma: %q", arg)
		}
		tags = append(tags, tag)
	}
	return tags, nil
}
