package tmux

type SetPaneOptionParams struct {
	TargetPane string
}

// SetPaneOption runs `set-option -p` to set a pane option.
func SetPaneOption(option, value string, params *SetPaneOptionParams) error {
	args := []string{"set-option", "-p"}
	if params.TargetPane != "" {
		args = append(args, "-t", params.TargetPane)
	}
	args = append(args, option, value)

	return run(args...)
}

// UnsetPaneOption runs `set-option -pu` to unset a pane option.
func UnsetPaneOption(option string, params *SetPaneOptionParams) error {
	args := []string{"set-option", "-pu"}
	if params.TargetPane != "" {
		args = append(args, "-t", params.TargetPane)
	}
	args = append(args, option)

	return run(args...)
}

type ResizePaneParams struct {
	TargetPane string
	ToggleZoom bool // -Z
}

// ResizePane runs `resize-pane`.
func ResizePane(params *ResizePaneParams) error {
	args := []string{"resize-pane"}
	if params.ToggleZoom {
		args = append(args, "-Z")
	}
	if params.TargetPane != "" {
		args = append(args, "-t", params.TargetPane)
	}

	return run(args...)
}

type ListPanesParams struct {
	Format string
	Target string
	All    bool // -a: list panes across every session and window
}

// ListPanes runs `list-panes` and returns the lines.
func ListPanes(params *ListPanesParams) ([]string, error) {
	args := []string{"list-panes"}
	if params.All {
		args = append(args, "-a")
	}
	if params.Format != "" {
		args = append(args, "-F", params.Format)
	}
	if params.Target != "" {
		args = append(args, "-t", params.Target)
	}

	return queryLines(args...)
}

type JoinPaneParams struct {
	SrcPane    string
	DstPane    string
	Detached   bool
	Vertical   bool
	Horizontal bool
	Before     bool
	Full       bool
	Size       string
}

// JoinPane runs `join-pane`.
func JoinPane(params *JoinPaneParams) error {
	args := []string{"join-pane"}
	if params.Before {
		args = append(args, "-b")
	}
	if params.Detached {
		args = append(args, "-d")
	}
	if params.Full {
		args = append(args, "-f")
	}
	if params.Horizontal {
		args = append(args, "-h")
	}
	if params.Vertical {
		args = append(args, "-v")
	}
	if params.Size != "" {
		args = append(args, "-l", params.Size)
	}
	if params.SrcPane != "" {
		args = append(args, "-s", params.SrcPane)
	}
	if params.DstPane != "" {
		args = append(args, "-t", params.DstPane)
	}

	return run(args...)
}

type DisplayPanesParams struct {
	Duration string // -d duration ("0" stays until a key is pressed)
	Template string // command template; %% is replaced by the selected pane id
}

// DisplayPanes runs `display-panes`, showing the numbered pane indicator and
// executing Template (with %% replaced by the chosen pane id) on selection.
func DisplayPanes(params *DisplayPanesParams) error {
	args := []string{"display-panes"}
	if params.Duration != "" {
		args = append(args, "-d", params.Duration)
	}
	if params.Template != "" {
		args = append(args, params.Template)
	}

	return run(args...)
}

type SplitWindowParams struct {
	TargetPane     string
	Size           string
	Vertical       bool
	Horizontal     bool
	Before         bool
	Detached       bool // -d: create the pane without switching focus to it
	Full           bool
	StartDirectory string
	Environment    []string // VAR=value pairs (-e), set in the new pane
	ShellCommand   []string // command (and args) to run; empty uses the default shell
	Format         string   // when set, print new pane info via -P -F <format>
}

// SplitWindow runs `split-window` to create a new pane. When Format is set it
// runs with -P -F and returns the formatted output; otherwise it returns "".
func SplitWindow(params *SplitWindowParams) (string, error) {
	args := []string{"split-window"}
	if params.Before {
		args = append(args, "-b")
	}
	if params.Detached {
		args = append(args, "-d")
	}
	if params.Full {
		args = append(args, "-f")
	}
	if params.Horizontal {
		args = append(args, "-h")
	}
	if params.Vertical {
		args = append(args, "-v")
	}
	if params.Format != "" {
		args = append(args, "-P", "-F", params.Format)
	}
	if params.Size != "" {
		args = append(args, "-l", params.Size)
	}
	if params.StartDirectory != "" {
		args = append(args, "-c", params.StartDirectory)
	}
	for _, env := range params.Environment {
		args = append(args, "-e", env)
	}
	if params.TargetPane != "" {
		args = append(args, "-t", params.TargetPane)
	}
	// The shell command and its arguments are positional and must come last.
	args = append(args, params.ShellCommand...)

	return query(args...)
}

type BreakPaneParams struct {
	SrcPane    string
	DstWindow  string
	Detached   bool
	WindowName string
}

// BreakPane runs `break-pane`.
func BreakPane(params *BreakPaneParams) error {
	args := []string{"break-pane"}
	if params.Detached {
		args = append(args, "-d")
	}
	if params.WindowName != "" {
		args = append(args, "-n", params.WindowName)
	}
	if params.SrcPane != "" {
		args = append(args, "-s", params.SrcPane)
	}
	if params.DstWindow != "" {
		args = append(args, "-t", params.DstWindow)
	}

	return run(args...)
}
