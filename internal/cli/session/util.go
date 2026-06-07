package session

import (
	"github.com/MunifTanjim/tmux-ctrl/internal/config"
	"github.com/MunifTanjim/tmux-ctrl/internal/tmux"
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
