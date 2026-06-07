package tui

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/MunifTanjim/tmux-ctrl/internal/shell"
	"github.com/MunifTanjim/tmux-ctrl/internal/util"
)

type FuzzySearchConfig[T any] struct {
	Header        string
	Items         []T
	AutoSelectOne bool
	Multi         bool
	Preview       func(T) string
	Query         string
}

type FuzzySearch[T any] struct {
	config FuzzySearchConfig[T]
}

func NewFuzzySearch[T any](conf FuzzySearchConfig[T]) *FuzzySearch[T] {
	if conf.Preview == nil {
		conf.Preview = func(item T) string {
			if s, ok := any(item).(fmt.Stringer); ok {
				return s.String()
			}
			return fmt.Sprintf("%v", item)
		}
	}
	return &FuzzySearch[T]{
		config: conf,
	}
}

var (
	ErrNoMatch   = errors.New("no match found")
	ErrCancelled = errors.New("cancelled")
)

func (f *FuzzySearch[T]) Run() ([]T, error) {
	if err := util.EnsureTool("fzf"); err != nil {
		return nil, err
	}

	args := []string{}

	if f.config.Header != "" {
		args = append(args, "--header", f.config.Header)
	}

	if f.config.AutoSelectOne {
		args = append(args, "--select-1")
	}

	if f.config.Multi {
		args = append(args, "--multi")
	}

	if f.config.Query != "" {
		args = append(args, "--query", f.config.Query)
	}

	items := make([]string, 0, len(f.config.Items))
	idxByPreview := make(map[string]int, len(f.config.Items))
	for i, item := range f.config.Items {
		preview := f.config.Preview(item)
		items = append(items, preview)
		idxByPreview[strings.TrimSpace(preview)] = i
	}

	cmd := shell.NewCommand("fzf", args...).
		WithStdIn(strings.NewReader(strings.Join(items, "\n"))).
		WithStdErr(os.Stderr)

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 130 {
				return nil, ErrCancelled
			}
			if exitErr.ExitCode() == 1 {
				return nil, ErrNoMatch
			}
		}
		return nil, fmt.Errorf("fzf failed: %w", err)
	}

	output := cmd.StdOut().TrimSpace().String()
	if output == "" {
		return nil, ErrCancelled
	}

	selected := strings.Split(output, "\n")
	result := make([]T, len(selected))
	for i, sel := range selected {
		if idx, ok := idxByPreview[strings.TrimSpace(sel)]; ok {
			result[i] = f.config.Items[idx]
		}
	}
	return result, nil
}
