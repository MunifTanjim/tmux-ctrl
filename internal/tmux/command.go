package tmux

import (
	"fmt"
	"strings"

	"github.com/MunifTanjim/tmux-ctrl/internal/shell"
)

// run executes a tmux subcommand, surfacing stderr on failure.
func run(args ...string) error {
	_, err := query(args...)
	return err
}

// query runs a tmux subcommand and returns trimmed stdout, surfacing stderr on failure.
func query(args ...string) (string, error) {
	out, err := queryRaw(args...)
	return strings.TrimSpace(out), err
}

// queryRaw is like query but returns stdout verbatim (e.g. keeping a captured
// pane's first-line indentation that query would trim).
func queryRaw(args ...string) (string, error) {
	cmd := shell.NewCommand("tmux", args...)
	if err := cmd.Run(); err != nil {
		if stderr := cmd.StdErr().TrimSpace().String(); stderr != "" {
			return "", fmt.Errorf("%s", stderr)
		}
		return "", err
	}
	return cmd.StdOut().String(), nil
}

// queryLines runs a tmux subcommand and returns its stdout split into lines.
func queryLines(args ...string) ([]string, error) {
	out, err := query(args...)
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	return strings.Split(out, "\n"), nil
}
