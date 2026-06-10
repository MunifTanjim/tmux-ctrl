package pane

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/MunifTanjim/tmux-ctrl/internal/config"
	"github.com/MunifTanjim/tmux-ctrl/internal/tmux"
	"github.com/MunifTanjim/tmux-ctrl/internal/tui"
	"github.com/spf13/cobra"
)

type match struct {
	Value string
	Type  string
	Line  int
	Col   int // rune column of the value's start within its line
	Label string
}

type namedPattern struct {
	name string
	re   *regexp.Regexp
}

// urlTrailingTrim is punctuation stripped from the end of a URL match (e.g. a
// URL printed inside parentheses or at the end of a sentence).
const urlTrailingTrim = `.,;:!?)]}>"'`

// builtinPatterns are tried in this order; the first to claim a value sets its type.
var builtinPatterns = []namedPattern{
	{name: "url", re: regexp.MustCompile(`(?:https?://|www\.)[^\s]+`)},
	{name: "path", re: regexp.MustCompile(`(?:~|\.{1,2})?/[\w.~/@%+-]+`)},
	{name: "sha", re: regexp.MustCompile(`\b[0-9a-f]{7,40}\b`)},
}

func ExtractCommand() *cobra.Command {
	var (
		paneID       string
		types        []string
		hint         bool
		copyToBuffer bool
	)

	cmd := &cobra.Command{
		Use:   "extract",
		Short: "Extract tokens (URLs, paths, SHAs, ...) from a pane",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			paneID, err := resolvePaneID(paneID)
			if err != nil {
				return err
			}

			pats, err := buildPatterns(types)
			if err != nil {
				return err
			}

			content, err := tmux.CapturePane(&tmux.CapturePaneParams{TargetPane: paneID, Join: true})
			if err != nil {
				return err
			}

			var values []string
			if hint {
				values, err = pickViaHint(content, pats)
			} else {
				values, err = pickViaFzf(content, pats)
			}
			if err != nil {
				return err
			}
			if len(values) == 0 {
				return nil
			}

			if copyToBuffer {
				return tmux.SetBuffer(strings.Join(values, "\n"))
			}
			for _, value := range values {
				tui.StdOutLn(value)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&paneID, "pane-id", "p", "", "pane to extract from (default: current pane)")
	cmd.Flags().StringSliceVar(&types, "type", nil, "limit to these pattern types (default: all)")
	cmd.Flags().BoolVar(&hint, "hint", false, "pick with an in-pane hint overlay instead of fzf")
	cmd.Flags().BoolVar(&copyToBuffer, "copy", false, "copy the selection to the tmux buffer instead of printing it")

	_ = cmd.RegisterFlagCompletionFunc("pane-id", paneCompletion(false))
	_ = cmd.RegisterFlagCompletionFunc("type", extractTypeCompletion)

	return cmd
}

// pickViaHint shows the in-pane hint overlay and returns the single chosen value
// (nil when there were no matches or the overlay was cancelled). It extracts from
// the joined content so wrapped tokens stay whole; the overlay re-wraps to the
// terminal width it sees so hints land on the right visual cell.
func pickViaHint(content string, pats []namedPattern) ([]string, error) {
	logical := strings.Split(content, "\n")
	matches := extractMatches(logical, pats)
	if len(matches) == 0 {
		tui.StdErrLn("No matches")
		return nil, nil
	}

	value, err := runHintOverlay(logical, matches)
	if err != nil || value == "" {
		return nil, err
	}
	return []string{value}, nil
}

// pickViaFzf shows the fzf picker and returns the chosen values (nil when there
// were no matches or the picker was cancelled).
func pickViaFzf(content string, pats []namedPattern) ([]string, error) {
	matches := extractMatches(strings.Split(content, "\n"), pats)
	if len(matches) == 0 {
		tui.StdErrLn("No matches")
		return nil, nil
	}

	selected, err := tui.NewFZFSearch(tui.FZFSearchConfig[match]{
		Header:        "Extract",
		Items:         matches,
		Multi:         true,
		AutoSelectOne: true,
		Fields: func(m match) []string {
			return []string{m.Value, m.Type}
		},
	}).Run()
	if err != nil {
		if errors.Is(err, tui.ErrCancelled) || errors.Is(err, tui.ErrNoMatch) {
			return nil, nil
		}
		return nil, err
	}

	values := make([]string, len(selected))
	for i, m := range selected {
		values[i] = m.Value
	}
	return values, nil
}

// extractMatches scans lines in priority order, returning the first occurrence of
// each distinct value with its position. A lower-priority pattern cannot match
// inside a span an earlier one already claimed, so e.g. the path pattern won't
// pull a fragment out of a URL.
func extractMatches(lines []string, pats []namedPattern) []match {
	var matches []match
	seen := make(map[string]bool)

	for lineIdx, line := range lines {
		var claimed [][2]int
		overlaps := func(start, end int) bool {
			for _, c := range claimed {
				if start < c[1] && c[0] < end {
					return true
				}
			}
			return false
		}

		for _, p := range pats {
			for _, loc := range p.re.FindAllStringIndex(line, -1) {
				start, end := loc[0], loc[1]
				if overlaps(start, end) {
					continue
				}
				value := line[start:end]
				if p.name == "url" {
					trimmed := strings.TrimRight(value, urlTrailingTrim)
					end -= len(value) - len(trimmed)
					value = trimmed
				}
				claimed = append(claimed, [2]int{start, end})
				if value == "" || seen[value] {
					continue
				}
				seen[value] = true
				matches = append(matches, match{
					Value: value,
					Type:  p.name,
					Line:  lineIdx,
					Col:   utf8.RuneCountInString(line[:start]),
				})
			}
		}
	}

	return matches
}

// buildPatterns returns the built-in patterns merged with config-defined custom
// ones (a custom name matching a built-in overrides it). When types is non-empty
// the result is restricted to those names, in the given order.
func buildPatterns(types []string) ([]namedPattern, error) {
	pats := append([]namedPattern(nil), builtinPatterns...)

	custom := config.Get[map[string]string]("extract.patterns")
	names := make([]string, 0, len(custom))
	for name := range custom {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		re, err := regexp.Compile(custom[name])
		if err != nil {
			return nil, fmt.Errorf("invalid regex for pattern %q: %w", name, err)
		}
		np := namedPattern{name: name, re: re}
		if i := indexPattern(pats, name); i >= 0 {
			pats[i] = np
		} else {
			pats = append(pats, np)
		}
	}

	if len(types) == 0 {
		return pats, nil
	}

	filtered := make([]namedPattern, 0, len(types))
	for _, name := range types {
		i := indexPattern(pats, name)
		if i < 0 {
			return nil, fmt.Errorf("unknown pattern type %q", name)
		}
		filtered = append(filtered, pats[i])
	}
	return filtered, nil
}

func indexPattern(pats []namedPattern, name string) int {
	return slices.IndexFunc(pats, func(p namedPattern) bool { return p.name == name })
}

func extractTypeCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	pats, err := buildPatterns(nil)
	if err != nil {
		pats = builtinPatterns
	}
	names := make([]string, len(pats))
	for i, p := range pats {
		names[i] = p.name
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
