package tmux

import (
	"strconv"
	"strings"
)

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

type DisplayMenuItem struct {
	Name    string // label shown in the menu
	Key     string // mnemonic key (underlined in Name by tmux)
	Command string // tmux command run when the item is chosen
}

type DisplayMenuParams struct {
	Title string
	Items []DisplayMenuItem
}

// DisplayMenu runs `display-menu`, showing a popup menu; the Command of the
// chosen item is executed by tmux on selection.
func DisplayMenu(params *DisplayMenuParams) error {
	args := []string{"display-menu"}
	if params.Title != "" {
		args = append(args, "-T", params.Title)
	}
	for _, item := range params.Items {
		if item.Name == "" {
			args = append(args, "") // separator line
			continue
		}
		args = append(args, item.Name, item.Key, item.Command)
	}

	return run(args...)
}

// menuKeys is the pool of single-character display-menu mnemonics, assigned in order.
const menuKeys = "123456789abcdefghijklmnopqrstuvwxyz"

// MenuKey returns the mnemonic key for the i-th menu item, or "" once the pool
// is exhausted.
func MenuKey(i int) string {
	if i < len(menuKeys) {
		return string(menuKeys[i])
	}
	return ""
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

type CapturePaneParams struct {
	TargetPane string
	Join       bool // -J: rejoin wrapped lines (so a wrapped URL stays one token)

	// StartLine/EndLine map to -S/-E: line 0 is the visible screen's top,
	// negatives reach into scrollback. Both nil captures the visible screen.
	StartLine *int
	EndLine   *int
}

func CapturePane(params *CapturePaneParams) (string, error) {
	args := []string{"capture-pane", "-p"}
	if params.Join {
		args = append(args, "-J")
	}
	if params.StartLine != nil {
		args = append(args, "-S", strconv.Itoa(*params.StartLine))
	}
	if params.EndLine != nil {
		args = append(args, "-E", strconv.Itoa(*params.EndLine))
	}
	if params.TargetPane != "" {
		args = append(args, "-t", params.TargetPane)
	}

	// queryRaw, not query: query's trim would drop the first line's indentation,
	// shifting every column and misaligning the hint overlay.
	out, err := queryRaw(args...)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(out, "\n"), nil
}

type SwapPaneParams struct {
	SrcPane  string // -s
	DstPane  string // -t
	Detached bool   // -d: keep the active pane (don't follow the swap)
}

// SwapPane runs `swap-pane`.
func SwapPane(params *SwapPaneParams) error {
	args := []string{"swap-pane"}
	if params.Detached {
		args = append(args, "-d")
	}
	if params.SrcPane != "" {
		args = append(args, "-s", params.SrcPane)
	}
	if params.DstPane != "" {
		args = append(args, "-t", params.DstPane)
	}

	return run(args...)
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
