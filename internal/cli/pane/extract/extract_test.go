package extract

import (
	"strings"
	"testing"
)

func TestExtractMatchesPositions(t *testing.T) {
	lines := []string{
		"open https://example.com/a here",
		"file /etc/hosts and sha a1b2c3d",
		"dup https://example.com/a again",
	}
	got := Match(lines, builtinPatterns)

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

func TestMatchSpans(t *testing.T) {
	eq := func(a, b []hintSpan) bool {
		if len(a) != len(b) {
			return false
		}
		for i := range a {
			if a[i] != b[i] {
				return false
			}
		}
		return true
	}

	// single row: "abc" (len 3) at col 2, width 10 -> one span.
	got := matchSpans([]match{{Value: "abc", Line: 0, Col: 2, Label: "a"}}, []int{0}, 10)
	want := []hintSpan{{line: 0, start: 2, end: 5, label: "a", labelHere: true}}
	if !eq(got, want) {
		t.Errorf("single row = %+v, want %+v", got, want)
	}

	// wrapping: len 8 at col 6, width 10 -> two spans across rows.
	got = matchSpans([]match{{Value: "12345678", Line: 0, Col: 6, Label: "b"}}, []int{0}, 10)
	want = []hintSpan{
		{line: 0, start: 6, end: 10, label: "b", labelHere: true},
		{line: 1, start: 0, end: 4, label: "b", labelHere: false},
	}
	if !eq(got, want) {
		t.Errorf("wrapping = %+v, want %+v", got, want)
	}

	// multi logical line: line 1 starts at visual row 3; col 23 len 2, width 10.
	got = matchSpans([]match{{Value: "ab", Line: 1, Col: 23, Label: "c"}}, []int{0, 3}, 10)
	want = []hintSpan{{line: 5, start: 3, end: 5, label: "c", labelHere: true}}
	if !eq(got, want) {
		t.Errorf("multi line = %+v, want %+v", got, want)
	}

	// width <= 0: one span on the logical line, unwrapped.
	got = matchSpans([]match{{Value: "abc", Line: 0, Col: 4, Label: "d"}}, []int{0}, 0)
	want = []hintSpan{{line: 0, start: 4, end: 7, label: "d", labelHere: true}}
	if !eq(got, want) {
		t.Errorf("width<=0 = %+v, want %+v", got, want)
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

func TestPreparePatterns(t *testing.T) {
	// Default order must be deterministic (Match relies on it for priority).
	a, err := PreparePatterns(nil)
	if err != nil {
		t.Fatalf("PreparePatterns(nil): %v", err)
	}
	if len(a) != len(builtinPatterns) || a[0].Name != "md-url" {
		t.Fatalf("unexpected default patterns, first=%q len=%d", a[0].Name, len(a))
	}
	b, _ := PreparePatterns(nil)
	for i := range a {
		if a[i].Name != b[i].Name {
			t.Fatalf("non-deterministic order at %d: %q vs %q", i, a[i].Name, b[i].Name)
		}
	}

	// Filtering keeps the requested order; unknown names error.
	got, err := PreparePatterns([]string{"sha", "url"})
	if err != nil || len(got) != 2 || got[0].Name != "sha" || got[1].Name != "url" {
		t.Errorf("filter = %v (err %v), want [sha url]", got, err)
	}
	if _, err := PreparePatterns([]string{"nope"}); err == nil {
		t.Error("expected error for unknown pattern name")
	}
}

func TestExtractMatchesValues(t *testing.T) {
	cases := []struct {
		name    string
		content string
		pats    []*extractPattern
		want    []match
	}{
		{
			name:    "url with trailing punctuation",
			content: "see (https://example.com/foo).",
			pats:    builtinPatterns,
			want:    []match{{Value: "https://example.com/foo", Type: "url"}},
		},
		{
			name:    "markdown link emits the inner url",
			content: "see [docs](https://example.com/x) ok",
			pats:    builtinPatterns,
			want:    []match{{Value: "https://example.com/x", Type: "md-url"}},
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
			name:    "bare relative path",
			content: "see src/main.go please",
			pats:    builtinPatterns,
			want:    []match{{Value: "src/main.go", Type: "path"}},
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
			pats:    []*extractPattern{{Name: "jira", Pattern: re(`[A-Z]+-[0-9]+`)}},
			want:    []match{{Value: "ABC-123", Type: "jira"}},
		},
		{
			name:    "custom pattern capture group is the value",
			content: "see Issue #42 now",
			pats:    []*extractPattern{{Name: "issue", Pattern: re(`Issue #([0-9]+)`)}},
			want:    []match{{Value: "42", Type: "issue"}},
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
			got := Match(strings.Split(c.content, "\n"), c.pats)
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
