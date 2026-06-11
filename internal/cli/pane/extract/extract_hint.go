package extract

import (
	"os"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/MunifTanjim/tmux-ctrl/internal/config"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const defaultHintAlphabet = "asdfqwerzxcvjklmiuopghtybn"

var (
	hintDimStyle   = lipgloss.NewStyle().Faint(true)
	hintTokenStyle = lipgloss.NewStyle() // undimmed, so the matched token stays readable
	hintKeyStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("3")).Bold(true)
	hintTypedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("2")).Bold(true)
)

// hintAlphabet returns the configured hint key pool, falling back to the default.
func hintAlphabet() string {
	if a := config.Get[string]("extract.hint_alphabet"); utf8.RuneCountInString(a) >= 2 {
		return a
	}
	return defaultHintAlphabet
}

// assignLabels gives each match a fixed-length hint label over alphabet and
// returns the label length (smallest k with len(alphabet)^k >= len(matches)).
func assignLabels(matches []match, alphabet string) int {
	alpha := []rune(alphabet)
	base := len(alpha)

	labelLen := 1
	for capacity := base; capacity < len(matches); capacity *= base {
		labelLen++
	}

	for i := range matches {
		matches[i].Label = encodeLabel(i, labelLen, alpha)
	}
	return labelLen
}

// encodeLabel renders index as a fixed-width base-len(alpha) string, MSB first.
func encodeLabel(index, length int, alpha []rune) string {
	base := len(alpha)
	buf := make([]rune, length)
	for pos := length - 1; pos >= 0; pos-- {
		buf[pos] = alpha[index%base]
		index /= base
	}
	return string(buf)
}

func resolveLabel(matches []match, label string) (string, bool) {
	if i := slices.IndexFunc(matches, func(m match) bool { return m.Label == label }); i >= 0 {
		return matches[i].Value, true
	}
	return "", false
}

// runHintOverlay shows the dimmed snapshot with hint labels and returns the
// chosen match's value ("" if cancelled). lines/matches are in logical
// coordinates; the model re-wraps to the terminal width via tea.WindowSizeMsg.
func runHintOverlay(lines []string, matches []match) (string, error) {
	labelLen := assignLabels(matches, hintAlphabet())
	m := hintOverlayModel{lines: lines, matches: matches, labelLen: labelLen}

	opts := []tea.ProgramOption{tea.WithAltScreen()}
	// stdout carries the picked value, so drive the overlay off the tty instead —
	// otherwise the UI vanishes into the pipe and lipgloss strips color.
	if tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err == nil {
		defer tty.Close()
		opts = append(opts, tea.WithInput(tty), tea.WithOutput(tty))
		lipgloss.SetColorProfile(lipgloss.NewRenderer(tty).ColorProfile())
	}

	final, err := tea.NewProgram(m, opts...).Run()
	if err != nil {
		return "", err
	}
	return final.(hintOverlayModel).selected, nil
}

type hintOverlayModel struct {
	lines    []string
	matches  []match
	labelLen int

	// derived from the terminal width (tea.WindowSizeMsg)
	renderedLines []string
	renderedSpans []hintSpan

	input    string
	selected string
	peek     bool // Space toggles this to hide hints and reveal the matches' text
}

func (m hintOverlayModel) Init() tea.Cmd { return nil }

func (m hintOverlayModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		renderedLines, startRowIndices := wrapLines(m.lines, msg.Width)
		m.renderedLines = renderedLines
		m.renderedSpans = matchSpans(m.matches, startRowIndices, msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeySpace:
			m.peek = !m.peek
			m.input = ""
			return m, nil
		case tea.KeyRunes:
			if m.peek {
				return m, nil
			}
			m.input += string(msg.Runes)
			if utf8.RuneCountInString(m.input) >= m.labelLen {
				if value, ok := resolveLabel(m.matches, m.input); ok {
					m.selected = value
				}
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m hintOverlayModel) View() string {
	// Before the first WindowSizeMsg, render the unwrapped snapshot without hints.
	lines, spans := m.renderedLines, m.renderedSpans
	if lines == nil {
		lines = m.lines
		spans = nil
	}

	// Keep only spans whose label still matches the typed prefix.
	spansByLine := make(map[int][]hintSpan)
	for _, s := range spans {
		if m.input == "" || strings.HasPrefix(s.label, m.input) {
			spansByLine[s.line] = append(spansByLine[s.line], s)
		}
	}

	var sb strings.Builder
	for i, line := range lines {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(renderHintLine(line, spansByLine[i], m.input, m.peek))
	}
	return sb.String()
}

// renderHintLine dims the line and undims each span; the hint label is drawn on
// the span where the match begins (omitted in peek mode).
func renderHintLine(line string, spans []hintSpan, input string, peek bool) string {
	if len(spans) == 0 {
		return hintDimStyle.Render(line)
	}
	slices.SortFunc(spans, func(a, b hintSpan) int { return a.start - b.start })

	runes := []rune(line)
	var sb strings.Builder
	col := 0
	for _, s := range spans {
		start, end := min(s.start, len(runes)), min(s.end, len(runes))
		if start > col {
			sb.WriteString(hintDimStyle.Render(string(runes[col:start])))
			col = start
		}
		if !peek && s.labelHere {
			sb.WriteString(renderLabel(s.label, input))
			col += utf8.RuneCountInString(s.label)
		}
		if end > col {
			sb.WriteString(hintTokenStyle.Render(string(runes[col:end])))
			col = end
		}
	}
	if col < len(runes) {
		sb.WriteString(hintDimStyle.Render(string(runes[col:])))
	}
	return sb.String()
}

func renderLabel(label, input string) string {
	if input != "" && strings.HasPrefix(label, input) {
		return hintTypedStyle.Render(label[:len(input)]) + hintKeyStyle.Render(label[len(input):])
	}
	return hintKeyStyle.Render(label)
}

// wrapLines re-wraps each logical line into visual rows of at most width runes,
// returning the visual lines and the first visual row index of each logical line.
// width <= 0 leaves lines unwrapped (one visual row per logical line).
func wrapLines(lines []string, width int) ([]string, []int) {
	startRowIndices := make([]int, len(lines))
	var rendered []string
	for i, line := range lines {
		startRowIndices[i] = len(rendered)
		if width <= 0 {
			rendered = append(rendered, line)
			continue
		}
		runes := []rune(line)
		if len(runes) == 0 {
			rendered = append(rendered, "")
			continue
		}
		for off := 0; off < len(runes); off += width {
			end := off + width
			if end > len(runes) {
				end = len(runes)
			}
			rendered = append(rendered, string(runes[off:end]))
		}
	}
	return rendered, startRowIndices
}

// hintSpan is a match's extent on a single visual row of the wrapped grid.
type hintSpan struct {
	line       int    // visual row
	start, end int    // rune column range [start,end) on that row
	label      string // the match's label (shared by all its spans)
	labelHere  bool   // draw the label on this span (the match's first row)
}

// matchSpans maps each match (in logical coordinates) onto the wrapped visual
// grid produced by wrapLines, splitting a match that spans rows into one span per
// row so every row it covers is undimmed. width <= 0 leaves lines unwrapped.
func matchSpans(matches []match, startRowIndices []int, width int) []hintSpan {
	var spans []hintSpan
	for _, m := range matches {
		base := startRowIndices[m.Line]
		length := utf8.RuneCountInString(m.Value)
		if length == 0 {
			continue
		}

		if width <= 0 {
			spans = append(spans, hintSpan{line: base, start: m.Col, end: m.Col + length, label: m.Label, labelHere: true})
			continue
		}

		first, last := m.Col, m.Col+length // logical columns [first, last)
		for c := first; c < last; {
			chunk := c / width
			rowStart := chunk * width
			segEnd := min(last, rowStart+width)
			spans = append(spans, hintSpan{
				line:      base + chunk,
				start:     c - rowStart,
				end:       segEnd - rowStart,
				label:     m.Label,
				labelHere: c == first,
			})
			c = segEnd
		}
	}
	return spans
}
