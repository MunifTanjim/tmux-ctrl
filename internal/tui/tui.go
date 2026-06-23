package tui

import (
	"fmt"
	"strings"

	"github.com/MunifTanjim/tmux-ctrl/internal/shell"
)

func Lines(a ...any) string {
	lines := make([]string, len(a))
	for i, v := range a {
		lines[i] = fmt.Sprint(v)
	}
	return strings.Join(lines, "\n")
}

func StdErr(a ...any) {
	shell.StdErr(a...)
}

func StdErrF(format string, a ...any) {
	shell.StdErrF(format, a...)
}

func StdErrLn(a ...any) {
	shell.StdErrLn(a...)
}

func StdOut(a ...any) {
	shell.StdOut(a...)
}

func StdOutF(format string, a ...any) {
	shell.StdOutF(format, a...)
}

func StdOutLn(a ...any) {
	shell.StdOutLn(a...)
}

func Exit(code int) {
	shell.Exit(code)
}
