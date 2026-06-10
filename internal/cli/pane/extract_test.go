package pane

import (
	"regexp"
	"strings"
	"testing"
)

func TestExtractMatchesPositions(t *testing.T) {
	lines := []string{
		"open https://example.com/a here",
		"file /etc/hosts and sha a1b2c3d",
		"dup https://example.com/a again",
	}
	got := extractMatches(lines, builtinPatterns)

	want := []match{
		{Value: "https://example.com/a", Type: "url", Line: 0, Col: 5},
		{Value: "/etc/hosts", Type: "path", Line: 1, Col: 5},
		{Value: "a1b2c3d", Type: "sha", Line: 1, Col: 24},
		// the duplicate URL on line 2 is dropped (dedup keeps first occurrence)
	}
	if len(got) != len(want) {
		t.Fatalf("extractMatches() = %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("match[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestAssignLabels(t *testing.T) {
	alpha := "abc" // base 3

	cases := []struct {
		n           int
		wantLen     int
		wantSamples map[int]string // index -> expected label
	}{
		{n: 1, wantLen: 1, wantSamples: map[int]string{0: "a"}},
		{n: 3, wantLen: 1, wantSamples: map[int]string{0: "a", 2: "c"}},
		{n: 4, wantLen: 2, wantSamples: map[int]string{0: "aa", 1: "ab", 3: "ba"}},
	}

	for _, c := range cases {
		matches := make([]match, c.n)
		labelLen := assignLabels(matches, alpha)
		if labelLen != c.wantLen {
			t.Errorf("assignLabels(n=%d) len = %d, want %d", c.n, labelLen, c.wantLen)
		}
		seen := map[string]bool{}
		for i, m := range matches {
			if len([]rune(m.Label)) != c.wantLen {
				t.Errorf("n=%d label[%d]=%q len %d, want %d", c.n, i, m.Label, len([]rune(m.Label)), c.wantLen)
			}
			if seen[m.Label] {
				t.Errorf("n=%d duplicate label %q", c.n, m.Label)
			}
			seen[m.Label] = true
		}
		for idx, want := range c.wantSamples {
			if matches[idx].Label != want {
				t.Errorf("n=%d label[%d] = %q, want %q", c.n, idx, matches[idx].Label, want)
			}
		}
	}
}

func TestWrapLines(t *testing.T) {
	// width 5: "0123456789ab" (12 runes) -> 3 rows; "" -> 1 row; "short" -> 1 row.
	lines := []string{"0123456789ab", "", "short"}
	visual, startRow := wrapLines(lines, 5)

	wantVisual := []string{"01234", "56789", "ab", "", "short"}
	if strings.Join(visual, "|") != strings.Join(wantVisual, "|") {
		t.Errorf("wrapLines visual = %q, want %q", visual, wantVisual)
	}
	wantStart := []int{0, 3, 4}
	for i := range wantStart {
		if startRow[i] != wantStart[i] {
			t.Errorf("startRow[%d] = %d, want %d", i, startRow[i], wantStart[i])
		}
	}

	// width <= 0 is identity.
	v, sr := wrapLines(lines, 0)
	if strings.Join(v, "|") != strings.Join(lines, "|") {
		t.Errorf("wrapLines(width=0) = %q, want identity", v)
	}
	for i := range sr {
		if sr[i] != i {
			t.Errorf("identity startRow[%d] = %d, want %d", i, sr[i], i)
		}
	}
}

func TestRemapMatches(t *testing.T) {
	// width 10, second logical line starts at visual row 1.
	// col 23 on logical line 1 -> +2 rows, col 3 -> visual row 3, col 3.
	startRow := []int{0, 1}
	matches := []match{{Line: 1, Col: 23}}
	remapMatches(matches, startRow, 10)
	if matches[0].Line != 3 || matches[0].Col != 3 {
		t.Errorf("remapMatches = (line %d, col %d), want (3, 3)", matches[0].Line, matches[0].Col)
	}

	// width <= 0 is a no-op.
	noop := []match{{Line: 1, Col: 23}}
	remapMatches(noop, startRow, 0)
	if noop[0].Line != 1 || noop[0].Col != 23 {
		t.Errorf("remapMatches(width=0) mutated match: %+v", noop[0])
	}
}

func TestResolveLabel(t *testing.T) {
	matches := []match{{Value: "x", Label: "aa"}, {Value: "y", Label: "ab"}}
	if v, ok := resolveLabel(matches, "ab"); !ok || v != "y" {
		t.Errorf("resolveLabel(ab) = (%q, %v), want (y, true)", v, ok)
	}
	if _, ok := resolveLabel(matches, "zz"); ok {
		t.Errorf("resolveLabel(zz) ok = true, want false")
	}
}

func TestExtractMatchesValues(t *testing.T) {
	cases := []struct {
		name    string
		content string
		pats    []namedPattern
		want    []match
	}{
		{
			name:    "url with trailing punctuation",
			content: "see (https://example.com/foo).",
			pats:    builtinPatterns,
			want:    []match{{Value: "https://example.com/foo", Type: "url"}},
		},
		{
			name:    "www url",
			content: "visit www.example.com, please",
			pats:    builtinPatterns,
			want:    []match{{Value: "www.example.com", Type: "url"}},
		},
		{
			name:    "paths absolute and relative",
			content: "edit /etc/hosts and ./src/main.go",
			pats:    builtinPatterns,
			want: []match{
				{Value: "/etc/hosts", Type: "path"},
				{Value: "./src/main.go", Type: "path"},
			},
		},
		{
			name:    "git shas short and full",
			content: "commit a1b2c3d then 0123456789abcdef0123456789abcdef01234567",
			pats:    builtinPatterns,
			want: []match{
				{Value: "a1b2c3d", Type: "sha"},
				{Value: "0123456789abcdef0123456789abcdef01234567", Type: "sha"},
			},
		},
		{
			name:    "dedup repeated value",
			content: "https://example.com/a\nhttps://example.com/a",
			pats:    builtinPatterns,
			want:    []match{{Value: "https://example.com/a", Type: "url"}},
		},
		{
			name:    "path pattern does not pull fragment out of url",
			content: "https://example.com/a/b",
			pats:    builtinPatterns,
			want:    []match{{Value: "https://example.com/a/b", Type: "url"}},
		},
		{
			name:    "custom pattern",
			content: "ticket ABC-123 ready",
			pats:    []namedPattern{{name: "jira", re: regexp.MustCompile(`[A-Z]+-[0-9]+`)}},
			want:    []match{{Value: "ABC-123", Type: "jira"}},
		},
		{
			name:    "no matches",
			content: "nothing interesting here",
			pats:    builtinPatterns,
			want:    nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := extractMatches(strings.Split(c.content, "\n"), c.pats)
			if len(got) != len(c.want) {
				t.Fatalf("extractMatches() = %v, want %v", got, c.want)
			}
			for i := range got {
				if got[i].Value != c.want[i].Value || got[i].Type != c.want[i].Type {
					t.Errorf("match[%d] = {%q, %q}, want {%q, %q}", i, got[i].Value, got[i].Type, c.want[i].Value, c.want[i].Type)
				}
			}
		})
	}
}
