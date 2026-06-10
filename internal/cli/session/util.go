package session

import (
	"errors"
	"strings"

	"github.com/MunifTanjim/tmux-ctrl/internal/config"
	"github.com/MunifTanjim/tmux-ctrl/internal/tmux"
	"github.com/MunifTanjim/tmux-ctrl/internal/tui"
	"github.com/MunifTanjim/tmux-ctrl/internal/util"
)

// listNames returns session names in tmux's order, excluding the hidden session.
func listNames() ([]string, error) {
	all, err := tmux.ListSessions(&tmux.ListSessionsParams{Format: "#{session_name}"})
	if err != nil {
		return nil, err
	}

	var names []string
	for _, name := range all {
		if name == config.HiddenSessionName {
			continue
		}
		names = append(names, name)
	}

	return names, nil
}

// switchRelative switches to the session offset positions away from the current
// one (wrapping), skipping the hidden session.
func switchRelative(offset int) error {
	names, err := listNames()
	if err != nil {
		return err
	}
	if len(names) == 0 {
		return nil
	}

	current, err := tmux.DisplayMessage("#{session_name}", &tmux.DisplayMessageParams{})
	if err != nil {
		return err
	}

	idx := -1
	for i, name := range names {
		if name == current {
			idx = i
			break
		}
	}

	// Current session is not in the visible list (e.g. attached to the hidden
	// session); jump to the first visible session.
	if idx < 0 {
		return tmux.SwitchClient(&tmux.SwitchClientParams{TargetSession: names[0]})
	}

	// Only the current visible session exists; nothing to switch to.
	if len(names) == 1 {
		return nil
	}

	target := names[(idx+offset+len(names))%len(names)]
	return tmux.SwitchClient(&tmux.SwitchClientParams{TargetSession: target})
}

type sessionInfo struct {
	Name    string
	Windows string
	Current bool
}

// listSessionInfos returns visible sessions (excluding the hidden one), marking the current one.
func listSessionInfos() ([]sessionInfo, error) {
	lines, err := tmux.ListSessions(&tmux.ListSessionsParams{
		Format: "#{session_name}\t#{session_windows}",
	})
	if err != nil {
		return nil, err
	}

	current, err := tmux.DisplayMessage("#{session_name}", &tmux.DisplayMessageParams{})
	if err != nil {
		return nil, err
	}

	infos := make([]sessionInfo, 0, len(lines))
	for _, line := range lines {
		fields := strings.SplitN(line, "\t", 2)
		if len(fields) < 2 {
			continue
		}
		name := fields[0]
		if name == config.HiddenSessionName {
			continue
		}
		infos = append(infos, sessionInfo{
			Name:    name,
			Windows: fields[1],
			Current: name == current,
		})
	}

	return infos, nil
}

func sessionSummary(info sessionInfo) string {
	summary := info.Windows + " windows"
	if info.Current {
		summary += " (current)"
	}
	return summary
}

// switchInteractive shows a picker and switches to the chosen session. It uses
// fzf when available and falls back to a tmux display-menu otherwise.
func switchInteractive() error {
	infos, err := listSessionInfos()
	if err != nil {
		return err
	}
	if len(infos) == 0 {
		return nil
	}

	if util.HasTool("fzf") {
		selected, err := tui.NewFZFSearch(tui.FZFSearchConfig[sessionInfo]{
			Header:        "Switch to session",
			Items:         infos,
			AutoSelectOne: true,
			Fields: func(info sessionInfo) []string {
				return []string{info.Name, sessionSummary(info)}
			},
			PreviewCommand: "tmux capture-pane -p -e -t '{1}'",
		}).Run()
		if err != nil {
			if errors.Is(err, tui.ErrCancelled) || errors.Is(err, tui.ErrNoMatch) {
				return nil
			}
			return err
		}
		if len(selected) == 0 {
			return nil
		}
		return tmux.SwitchClient(&tmux.SwitchClientParams{TargetSession: selected[0].Name})
	}

	return pickSessionMenu(infos)
}

// pickSessionMenu shows a display-menu of sessions, switching to the chosen one.
func pickSessionMenu(infos []sessionInfo) error {
	items := make([]tmux.DisplayMenuItem, 0, len(infos))
	for i, info := range infos {
		items = append(items, tmux.DisplayMenuItem{
			Name:    info.Name,
			Key:     tmux.MenuKey(i),
			Command: "switch-client -t " + info.Name,
		})
	}

	return tmux.DisplayMenu(&tmux.DisplayMenuParams{Title: "Switch session", Items: items})
}
