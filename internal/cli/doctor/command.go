package doctor

import (
	"fmt"
	"os"
	"strings"

	"github.com/MunifTanjim/tmux-ctrl/internal/config"
	"github.com/MunifTanjim/tmux-ctrl/internal/health"
	"github.com/MunifTanjim/tmux-ctrl/internal/shell"
	"github.com/MunifTanjim/tmux-ctrl/internal/tui"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	checkMark    = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).SetString("✓")
	crossMark    = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).SetString("✗")
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
)

func formatReason(reason string) string {
	if reason == "" {
		return ""
	}
	return fmt.Sprintf(" (%s)", reason)
}

func checkPassed(c health.Check) bool {
	if !c.Passed {
		return false
	}
	for _, meta := range c.Meta {
		if !meta.Passed {
			return false
		}
	}
	return true
}

type checkResult struct {
	index int
	check health.Check
}

type healthModel struct {
	checkers       []health.Checker
	results        []health.Check
	completed      []bool
	completedCount int
	spinner        spinner.Model
	allPassed      bool
}

func newHealthModel() healthModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	checkers := []health.Checker{}
	checkers = append(checkers, config.Health()...)
	checkers = append(checkers, shell.Health()...)

	return healthModel{
		checkers:  checkers,
		results:   make([]health.Check, len(checkers)),
		completed: make([]bool, len(checkers)),
		spinner:   s,
		allPassed: true,
	}
}

func runCheck(idx int, checker health.Checker) tea.Cmd {
	return func() tea.Msg {
		return checkResult{index: idx, check: checker.Check()}
	}
}

func (m healthModel) Init() tea.Cmd {
	cmds := make([]tea.Cmd, len(m.checkers)+1)
	cmds[0] = m.spinner.Tick
	for i, checker := range m.checkers {
		cmds[i+1] = runCheck(i, checker)
	}
	return tea.Batch(cmds...)
}

func (m healthModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case checkResult:
		m.results[msg.index] = msg.check
		m.completed[msg.index] = true
		m.completedCount++

		if !checkPassed(msg.check) {
			m.allPassed = false
		}

		if m.completedCount >= len(m.checkers) {
			return m, tea.Quit
		}
		return m, nil
	}

	return m, nil
}

func (m healthModel) View() string {
	var sb strings.Builder
	sb.WriteString("Health Check\n\n")

	for i, checker := range m.checkers {
		if m.completed[i] {
			sb.WriteString(renderCheck(checker, m.results[i]))
		} else {
			fmt.Fprintf(&sb, "  %s %s\n\n", m.spinner.View(), checker.Name)
		}
	}

	return sb.String()
}

func renderCheck(checker health.Checker, c health.Check) string {
	var sb strings.Builder
	mark := checkMark
	if !c.Passed {
		mark = crossMark
	}
	fmt.Fprintf(&sb, "  %s %s%s\n", mark, checker.Name, formatReason(c.Reason))
	for _, m := range c.Meta {
		fmt.Fprintf(&sb, "      %s: %s%s\n", m.Name, m.Value, formatReason(m.Reason))
	}
	sb.WriteString("\n")

	return sb.String()
}

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Run health checks",
		Long:  `Check the health of required tools, configurations, and services.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			m := newHealthModel()
			p := tea.NewProgram(m, tea.WithAltScreen())
			finalModel, err := p.Run()
			if err != nil {
				tui.StdErrF("Error: %v\n", err)
				os.Exit(1)
			}

			if fm, ok := finalModel.(healthModel); ok {
				tui.StdOut(fm.View())
				if !fm.allPassed {
					os.Exit(1)
				}
			}
		},
	}
}
