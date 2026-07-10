package extract

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/MunifTanjim/tmux-ctrl/internal/config"
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

type Regexp struct {
	*regexp.Regexp
}

func (r *Regexp) UnmarshalText(b []byte) error {
	regex, err := regexp.Compile(string(b))
	if err != nil {
		return err
	}
	r.Regexp = regex
	return nil
}

type extractPattern struct {
	Name    string  `mapstructure:"name"`
	Pattern *Regexp `mapstructure:"pattern"`
}

func re(pattern string) *Regexp {
	return &Regexp{regexp.MustCompile(pattern)}
}

// urlTrailingTrim is punctuation stripped from the end of a URL match (e.g. a
// URL printed inside parentheses or at the end of a sentence).
const urlTrailingTrim = `.,;:!?)]}>"'`

// builtinPatterns are tried in order; the first to claim a span sets the type.
var builtinPatterns = []*extractPattern{
	// command: capture group 1 (prompt stripped) is the value. Claims the whole
	// line first so tokens inside aren't pulled out. Optional leading status code
	// before the marker (e.g. starship's exit code before ❯).
	{Name: "command", Pattern: re(`^\s*(?:\d+\s+)?[$%❯]\s+(.*\S)`)},
	// md-url precedes url to claim the whole `[text](url)` and emit the inner URL
	// (capture group 1).
	{Name: "md-url", Pattern: re(`\[[^]]*\]\(([^)]+)\)`)},
	{Name: "url", Pattern: re(`(?:(?:file|ftp|git|https?|ssh)://|git@)[^\s]+`)},
	// quoted-path: quotes delimit the bounds, so interior spaces are safe; capture
	// group 1 (the path without quotes) is the value.
	{Name: "quoted-path", Pattern: re(`["']([^"']*/[^"']*)["']`)},
	// path: a backslash-escaped space (\ ) is part of a segment, so shell-style
	// paths with spaces match without over-running into trailing text.
	{Name: "path", Pattern: re(`(?:[~$\w.+-]|\\ )*(?:\/(?:[\w.+-]|\\ )+)+`)},
	// datetime precedes date to claim the whole span; both precede ip/ipv6/number
	// so the time isn't read as an address and the year isn't pulled out alone.
	{Name: "datetime", Pattern: re(`\b\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}(?::\d{2})?(?:\.\d+)?(?:Z|[+-]\d{2}:?\d{2})?`)},
	{Name: "date", Pattern: re(`\b\d{4}-\d{2}-\d{2}\b`)},
	// uuid precedes sha so sha doesn't claim the leading 8-hex chunk on its own.
	{Name: "uuid", Pattern: re(`\b[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}\b`)},
	{Name: "sha", Pattern: re(`\b[0-9a-f]{7,40}\b`)},
	{Name: "hex-color", Pattern: re(`\b#(?:[0-9a-f]{6}|[0-9A-F]{6})\b`)},
	{Name: "ip", Pattern: re(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)},
	{Name: "ipv6", Pattern: re(`[A-f0-9:]+:+[A-f0-9:]+[%\w\d]+`)},
	{Name: "number", Pattern: re(`[0-9]{4,}`)},
}

type Picker string

const (
	PickerHintOverlay Picker = "hint-overlay"
	PickerSelect      Picker = "select"
)

func Pick(picker Picker, lines []string, match []match) ([]string, error) {
	switch picker {
	case PickerHintOverlay:
		return pickUsingHintOverlay(lines, match)
	case PickerSelect:
		return pickUsingSelect(lines, match)
	default:
		return nil, fmt.Errorf("unknown picker: %s", picker)
	}
}

// pickUsingHintOverlay shows the hint overlay and returns the chosen value (nil if cancelled).
func pickUsingHintOverlay(lines []string, matches []match) ([]string, error) {
	value, err := runHintOverlay(lines, matches)
	if err != nil || value == "" {
		return nil, err
	}
	return []string{value}, nil
}

// pickUsingSelect shows the fzf picker and returns the chosen values (nil if cancelled).
func pickUsingSelect(lines []string, matches []match) ([]string, error) {
	selected, err := tui.NewFZFSearch(tui.FZFSearchConfig[match]{
		Header:        "Extract",
		Items:         matches,
		Multi:         true,
		AutoSelectOne: true,
		Fields: func(m match) []string {
			return []string{m.Value}
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

// Match scans lines in priority order, returning the first occurrence of
// each distinct value with its position. A lower-priority pattern cannot match
// inside a span an earlier one already claimed, so e.g. the path pattern won't
// pull a fragment out of a URL.
func Match(lines []string, pats []*extractPattern) []match {
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
			for _, loc := range p.Pattern.FindAllStringSubmatchIndex(line, -1) {
				start, end := loc[0], loc[1]
				if overlaps(start, end) {
					continue
				}

				// The value is the whole match, unless the pattern has a capture
				// group that participated — then its first group is the value
				// (e.g. md-url emits the URL inside `[text](url)`).
				valueStart, valueEnd := start, end
				if p.Pattern.NumSubexp() >= 1 && loc[2] >= 0 {
					valueStart, valueEnd = loc[2], loc[3]
				}

				value := line[valueStart:valueEnd]
				if p.Name == "url" {
					value = strings.TrimRight(value, urlTrailingTrim)
				}

				claimed = append(claimed, [2]int{start, end})
				if value == "" || seen[value] {
					continue
				}
				seen[value] = true
				matches = append(matches, match{
					Value: value,
					Type:  p.Name,
					Line:  lineIdx,
					Col:   utf8.RuneCountInString(line[:valueStart]),
				})
			}
		}
	}

	return matches
}

// PreparePatterns returns the built-in patterns merged with config-defined custom
// ones (a custom name matching a built-in overrides it, keeping its position).
// When patternNames is non-empty the result is restricted to those names, in the
// given order. Order is deterministic — Match relies on it for span priority.
func PreparePatterns(patternNames []string) ([]*extractPattern, error) {
	custom := config.Get[[]*extractPattern]("extract.patterns")

	merged := append([]*extractPattern{}, builtinPatterns...)
	for _, cp := range custom {
		if cp == nil || cp.Pattern == nil {
			continue
		}
		if i := slices.IndexFunc(merged, func(p *extractPattern) bool { return p.Name == cp.Name }); i >= 0 {
			merged[i] = cp
		} else {
			merged = append(merged, cp)
		}
	}

	if len(patternNames) == 0 {
		return merged, nil
	}

	patterns := make([]*extractPattern, 0, len(patternNames))
	for _, name := range patternNames {
		i := slices.IndexFunc(merged, func(p *extractPattern) bool { return p.Name == name })
		if i < 0 {
			return nil, fmt.Errorf("unknown pattern: %s", name)
		}
		patterns = append(patterns, merged[i])
	}
	return patterns, nil
}

func PatternNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	pats, err := PreparePatterns(nil)
	if err != nil {
		pats = builtinPatterns
	}
	names := make([]string, len(pats))
	for i, p := range pats {
		names[i] = p.Name
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
