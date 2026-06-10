package pane

import (
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/MunifTanjim/tmux-ctrl/internal/config"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const defaultHintAlphabet = "asdfghjklqwertyuiopzxcvbnm"

var (
	hintDimStyle   = lipgloss.NewStyle().Faint(true)
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
	for _, m := range matches {
		if m.Label == label {
			return m.Value, true
		}
	}
	return "", false
}

// runHintOverlay shows the dimmed snapshot with hint labels and returns the
// value of the chosen match ("" if cancelled). lines and matches are in logical
// (unwrapped) coordinates; the model re-wraps to the terminal width it receives
// via tea.WindowSizeMsg.
func runHintOverlay(lines []string, matches []match) (string, error) {
	labelLen := assignLabels(matches, hintAlphabet())
	m := hintModel{logicalLines: lines, logicalMatches: matches, labelLen: labelLen}

	final, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		return "", err
	}
	return final.(hintModel).selected, nil
}

type hintModel struct {
	logicalLines   []string
	logicalMatches []match
	labelLen       int

	// derived from the terminal width (tea.WindowSizeMsg)
	visualLines   []string
	visualMatches []match

	input    string
	selected string
}

func (m hintModel) Init() tea.Cmd { return nil }

func (m hintModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		visual, startRow := wrapLines(m.logicalLines, msg.Width)
		remapped := make([]match, len(m.logicalMatches))
		copy(remapped, m.logicalMatches)
		remapMatches(remapped, startRow, msg.Width)
		m.visualLines, m.visualMatches = visual, remapped
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyRunes:
			m.input += string(msg.Runes)
			if utf8.RuneCountInString(m.input) >= m.labelLen {
				if value, ok := resolveLabel(m.logicalMatches, m.input); ok {
					m.selected = value
				}
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m hintModel) View() string {
	// Before the first WindowSizeMsg, render the unwrapped snapshot without hints.
	lines, matches := m.visualLines, m.visualMatches
	if lines == nil {
		lines = m.logicalLines
		matches = nil
	}

	// Only matches whose label still matches what's been typed remain visible.
	byLine := make(map[int][]match)
	for _, mt := range matches {
		if m.input == "" || strings.HasPrefix(mt.Label, m.input) {
			byLine[mt.Line] = append(byLine[mt.Line], mt)
		}
	}

	var sb strings.Builder
	for i, line := range lines {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(renderHintLine(line, byLine[i], m.input))
	}
	return sb.String()
}

// renderHintLine draws line dimmed, with each match's label painted over the
// first labelLen runes of the match (preserving column width).
func renderHintLine(line string, matches []match, input string) string {
	if len(matches) == 0 {
		return hintDimStyle.Render(line)
	}
	slices.SortFunc(matches, func(a, b match) int { return a.Col - b.Col })

	runes := []rune(line)
	var sb strings.Builder
	col, mi := 0, 0
	for col < len(runes) {
		if mi < len(matches) && matches[mi].Col == col {
			label := matches[mi].Label
			sb.WriteString(renderLabel(label, input))
			col += utf8.RuneCountInString(label)
			mi++
			continue
		}

		next := len(runes)
		if mi < len(matches) && matches[mi].Col > col {
			next = matches[mi].Col
		}
		sb.WriteString(hintDimStyle.Render(string(runes[col:next])))
		col = next
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
	startRow := make([]int, len(lines))
	var visual []string
	for i, line := range lines {
		startRow[i] = len(visual)
		if width <= 0 {
			visual = append(visual, line)
			continue
		}
		runes := []rune(line)
		if len(runes) == 0 {
			visual = append(visual, "")
			continue
		}
		for off := 0; off < len(runes); off += width {
			end := off + width
			if end > len(runes) {
				end = len(runes)
			}
			visual = append(visual, string(runes[off:end]))
		}
	}
	return visual, startRow
}

// remapMatches rewrites each match's Line/Col from logical coordinates to the
// visual grid produced by wrapLines. A no-op when width <= 0.
func remapMatches(matches []match, startRow []int, width int) {
	if width <= 0 {
		return
	}
	for i := range matches {
		line, col := matches[i].Line, matches[i].Col
		matches[i].Line = startRow[line] + col/width
		matches[i].Col = col % width
	}
}
