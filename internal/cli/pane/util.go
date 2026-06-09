package pane

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/MunifTanjim/tmux-ctrl/internal/config"
	"github.com/MunifTanjim/tmux-ctrl/internal/tmux"
)

// hidden_session_name is the holding session where hidden panes are stashed.
const hidden_session_name = config.HiddenSessionName

// @-prefixed tmux options used as channels between tmux and this CLI.
const (
	// paneTagsOption holds a pane's comma-separated tags (see `pane tag`).
	paneTagsOption = "@tmux_ctrl_pane_tags"
	// paneMoveTargetOption ferries the pane id chosen in the display-panes picker
	// back to this process. Keeping pane ids out of the display-panes template
	// avoids corruption: display-panes only substitutes the `%%` token and
	// otherwise mangles stray `%` characters (and pane ids start with `%`).
	paneMoveTargetOption = "@tmux_ctrl_pane_move_target"
)

// hiddenPane describes a pane stashed in the hidden holding session.
type hiddenPane struct {
	Ref     string   // "<window_index>.<pane_index>" within the hidden session
	Command string   // pane_current_command
	Path    string   // pane_current_path
	Tags    []string // @tmux_ctrl_pane_tags recorded on the pane
}

// currentWindowLocation returns `<session-name>:<window-id>` for the active window.
func currentWindowLocation() (string, error) {
	return tmux.DisplayMessage("#{session_name}:#{window_id}", &tmux.DisplayMessageParams{})
}

// resolvePaneID returns paneID unchanged, or the current pane's id when empty.
func resolvePaneID(paneID string) (string, error) {
	if paneID != "" {
		return paneID, nil
	}
	return tmux.DisplayMessage("#{pane_id}", &tmux.DisplayMessageParams{})
}

// isPaneHidden reports whether paneID currently lives in the hidden session.
func isPaneHidden(paneID string) (bool, error) {
	session, err := tmux.DisplayMessage("#{session_name}", &tmux.DisplayMessageParams{TargetPane: paneID})
	if err != nil {
		return false, err
	}
	if session == "" {
		return false, errors.New("pane not found")
	}
	return session == hidden_session_name, nil
}

// hiddenWindowIndex returns the index of the window in the hidden session
// whose name matches location.
func hiddenWindowIndex(windowLoc string) (string, error) {
	lines, err := tmux.ListWindows(&tmux.ListWindowsParams{
		TargetSession: hidden_session_name,
		Format:        "#{window_index}\t#{window_name}",
	})
	if err != nil {
		return "", err
	}
	for _, line := range lines {
		if index, name, found := strings.Cut(line, "\t"); found && name == windowLoc {
			return index, nil
		}
	}
	return "", nil
}

// hiddenPaneLines runs list-panes with the given format against the hidden window
// for winLoc, returning the raw lines. Returns nil when the hidden session or its
// window for winLoc does not exist.
func hiddenPaneLines(winLoc, format string) ([]string, error) {
	if !tmux.HasSession(&tmux.HasSessionParams{TargetSession: hidden_session_name}) {
		return nil, nil
	}

	winIdx, err := hiddenWindowIndex(winLoc)
	if err != nil {
		return nil, err
	}

	if winIdx == "" {
		return nil, nil
	}

	return tmux.ListPanes(&tmux.ListPanesParams{
		Target: hidden_session_name + ":" + winIdx,
		Format: format,
	})
}

// listHiddenPanes returns the panes stashed for the given origin location.
func listHiddenPanes(winLoc string) ([]hiddenPane, error) {
	// Tags last: it is the only variable-length field, so keeping it at the end
	// avoids any tab in the value disturbing the earlier columns.
	lines, err := hiddenPaneLines(winLoc, "#{window_index}.#{pane_index}\t#{pane_current_command}\t#{pane_current_path}\t#{"+paneTagsOption+"}")
	if err != nil {
		return nil, err
	}

	panes := make([]hiddenPane, 0, len(lines))
	for _, line := range lines {
		fields := strings.SplitN(line, "\t", 4)
		if len(fields) == 3 {
			fields = append(fields, "")
		}
		if len(fields) < 4 {
			continue
		}
		panes = append(panes, hiddenPane{
			Ref:     fields[0],
			Command: fields[1],
			Path:    fields[2],
			Tags:    parseTags(fields[3]),
		})
	}

	return panes, nil
}

// hidePane moves a pane out of the visible layout into the hidden holding
// session, grouping hidden panes into a window named after the origin location
// so they can be restored later. paneID defaults to the current pane; a non-empty
// tag is added to the pane (deduped) before it is hidden.
func hidePane(paneID, tag string) error {
	winLoc, err := currentWindowLocation()
	if err != nil {
		return err
	}

	paneID, err = resolvePaneID(paneID)
	if err != nil {
		return err
	}

	if tag != "" {
		if err := addPaneTags(paneID, []string{tag}); err != nil {
			return err
		}
	}

	if !tmux.HasSession(&tmux.HasSessionParams{TargetSession: hidden_session_name}) {
		if err := tmux.NewSession(&tmux.NewSessionParams{SessionName: hidden_session_name, Detached: true}); err != nil {
			return err
		}
	}

	winIdx, err := hiddenWindowIndex(winLoc)
	if err != nil {
		return err
	}

	if winIdx != "" {
		return tmux.JoinPane(&tmux.JoinPaneParams{
			SrcPane:  paneID,
			DstPane:  hidden_session_name + ":" + winIdx,
			Detached: true,
		})
	}

	return tmux.BreakPane(&tmux.BreakPaneParams{
		SrcPane:    paneID,
		DstWindow:  hidden_session_name,
		Detached:   true,
		WindowName: winLoc,
	})
}

// ensureWindowUnzoomed unzooms target's window (empty = current) when zoomed, so
// later layout-geometry reads reflect the real pane arrangement. No-op otherwise.
func ensureWindowUnzoomed(targetPane string) error {
	flag, err := tmux.DisplayMessage("#{window_zoomed_flag}", &tmux.DisplayMessageParams{TargetPane: targetPane})
	if err != nil {
		return err
	}
	if flag != "1" {
		return nil
	}
	return tmux.ResizePane(&tmux.ResizePaneParams{TargetPane: targetPane, ToggleZoom: true})
}

// paneDirections lists every placement direction accepted by the show/move
// commands, in the order they are offered for completion.
var paneDirections = []string{
	"bottom", "top", "right", "left",
	"top-left", "top-right", "bottom-left", "bottom-right",
}

// directionsHint is the comma-separated direction list used in flag help text.
var directionsHint = strings.Join(paneDirections, ", ")

// paneSplitFlags maps a direction (bottom/top/right/left) to join-pane split
// orientation: whether the split is vertical/horizontal and whether the source
// is placed before (above/left of) the target.
func paneSplitFlags(direction string) (vertical, horizontal, before bool, err error) {
	switch direction {
	case "bottom":
		vertical = true
	case "top":
		vertical = true
		before = true
	case "right":
		horizontal = true
	case "left":
		horizontal = true
		before = true
	default:
		err = fmt.Errorf("invalid direction: %s (must be bottom, top, right, left)", direction)
	}
	return
}

// maxSplitSize returns the largest new-pane size that fits within avail cells,
// reserving one cell for the divider and one for the other pane. Returns 0 when
// no split fits.
func maxSplitSize(avail int) int {
	if avail < 3 {
		return 0
	}
	return avail - 2
}

// fractionToPercent converts a bare fraction 0<f<1 into a rounded percentage
// string (e.g. "0.3" -> "30%"). ok is false when size is not such a fraction.
func fractionToPercent(size string) (string, bool) {
	f, err := strconv.ParseFloat(size, 64)
	if err != nil || f <= 0 || f >= 1 {
		return "", false
	}
	return strconv.Itoa(int(math.Round(f*100))) + "%", true
}

// resolvePaneSize normalizes a split/join size for targetPane: a bare fraction
// 0<N<1 becomes a percentage, and an oversized absolute size is clamped so the
// other pane keeps at least one cell. size is returned unchanged when empty, a
// percentage, or a non-integer. vertical picks height vs width; full uses the
// window dimension (--edge/-f) instead of the target pane's dimension.
func resolvePaneSize(targetPane, size string, vertical, full bool) (string, error) {
	if size == "" || strings.HasSuffix(size, "%") {
		return size, nil
	}
	if pct, ok := fractionToPercent(size); ok {
		return pct, nil // percentages are bounded by tmux; no clamp needed
	}
	n, err := strconv.Atoi(size)
	if err != nil {
		return size, nil // not a plain integer; leave it to tmux
	}

	var dim string
	switch {
	case vertical && full:
		dim = "#{window_height}"
	case vertical:
		dim = "#{pane_height}"
	case full:
		dim = "#{window_width}"
	default:
		dim = "#{pane_width}"
	}

	raw, err := tmux.DisplayMessage(dim, &tmux.DisplayMessageParams{TargetPane: targetPane})
	if err != nil {
		return "", err
	}
	avail, err := strconv.Atoi(raw)
	if err != nil {
		return size, nil // can't determine; don't interfere
	}

	if max := maxSplitSize(avail); max >= 1 && n > max {
		return strconv.Itoa(max), nil
	}
	return size, nil
}

// splitOptions captures the inputs shared by the pane split helpers.
type splitOptions struct {
	paneID    string
	direction string
	size      string
	template  string   // tmux format string; when set, the new pane info is printed (-P -F)
	command   []string // command (and args) to run; empty uses the default shell
	env       []string // VAR=value variables to set in the new pane
	focus     bool     // when false, create the pane without switching focus (-d)
}

// splitPane creates a new pane by splitting o.paneID in o.direction
// (bottom/top/right/left) with an optional size. When o.command is non-empty it
// is run in the new pane instead of the default shell; o.env sets VAR=value
// variables. The new pane opens in the splitting pane's current directory and
// becomes active. When o.template is set, the rendered pane info is returned.
func splitPane(o splitOptions) (string, error) {
	vertical, horizontal, before, err := paneSplitFlags(o.direction)
	if err != nil {
		return "", err
	}

	size, err := resolvePaneSize(o.paneID, o.size, vertical, false)
	if err != nil {
		return "", err
	}

	return tmux.SplitWindow(&tmux.SplitWindowParams{
		TargetPane:     o.paneID,
		Size:           size,
		Vertical:       vertical,
		Horizontal:     horizontal,
		Before:         before,
		Detached:       !o.focus,
		StartDirectory: "#{pane_current_path}",
		Environment:    o.env,
		ShellCommand:   o.command,
		Format:         o.template,
	})
}

// splitPaneToEdge creates a new pane spanning the full window edge in o.direction
// (full width for top/bottom, full height for left/right) via split-window -f. A
// corner direction instead docks the new pane into that corner. See splitPane for
// o.command/o.env/o.template behavior.
func splitPaneToEdge(o splitOptions) (string, error) {
	if isCornerDirection(o.direction) {
		return splitPaneToCorner(o)
	}

	vertical, horizontal, before, err := paneSplitFlags(o.direction)
	if err != nil {
		return "", err
	}

	size, err := resolvePaneSize(o.paneID, o.size, vertical, true)
	if err != nil {
		return "", err
	}

	return tmux.SplitWindow(&tmux.SplitWindowParams{
		TargetPane:     o.paneID,
		Size:           size,
		Vertical:       vertical,
		Horizontal:     horizontal,
		Before:         before,
		Detached:       !o.focus,
		Full:           true,
		StartDirectory: "#{pane_current_path}",
		Environment:    o.env,
		ShellCommand:   o.command,
		Format:         o.template,
	})
}

// splitPaneToCorner docks a new pane into the corner of the current window named
// by o.direction, splitting the corner-most edge pane within its column (above
// for top-*, below for bottom-*). See splitPane for o.command/o.env/o.template
// behavior.
func splitPaneToCorner(o splitOptions) (string, error) {
	count, err := windowPaneCount("")
	if err != nil {
		return "", err
	}
	if count <= 1 {
		// One pane can't form a real corner from a single split; honor the corner's
		// horizontal side (left/right) instead of stacking full-width top/bottom.
		o.direction = cornerHorizontalDirection(o.direction)
		return splitPane(o)
	}

	anchor, joinDir, err := cornerAnchorPane("", o.direction, "")
	if err != nil {
		return "", err
	}
	if anchor == "" {
		return "", nil
	}

	o.paneID = anchor
	o.direction = joinDir
	return splitPane(o)
}

// joinPaneInDirection joins srcPane into dstPane, split according to direction
// (bottom/top/right/left) and an optional size. When full is set, the joined pane
// spans the whole window dimension (join-pane -f) rather than dstPane's cell.
func joinPaneInDirection(srcPane, dstPane, direction, size string, full bool) error {
	vertical, horizontal, before, err := paneSplitFlags(direction)
	if err != nil {
		return err
	}

	size, err = resolvePaneSize(dstPane, size, vertical, full)
	if err != nil {
		return err
	}

	return tmux.JoinPane(&tmux.JoinPaneParams{
		SrcPane:    srcPane,
		DstPane:    dstPane,
		Size:       size,
		Vertical:   vertical,
		Horizontal: horizontal,
		Before:     before,
		Full:       full,
	})
}

// cornerPlacement describes a corner direction in terms of the window edge to
// search for an anchor pane and how the moved pane is joined relative to it.
type cornerPlacement struct {
	atRight bool // search the right window edge instead of the left
	bottom  bool // anchor is the bottom-most pane on the edge; join below (else top-most; join above)
}

// corners maps each corner direction to its placement.
var corners = map[string]cornerPlacement{
	"top-left":     {atRight: false, bottom: false},
	"top-right":    {atRight: true, bottom: false},
	"bottom-left":  {atRight: false, bottom: true},
	"bottom-right": {atRight: true, bottom: true},
}

// isCornerDirection reports whether direction names a corner placement.
func isCornerDirection(direction string) bool {
	_, ok := corners[direction]
	return ok
}

// cornerHorizontalDirection returns the horizontal component (left/right) of a
// corner direction.
func cornerHorizontalDirection(direction string) string {
	if corners[direction].atRight {
		return "right"
	}
	return "left"
}

// cornerAnchorPane resolves the anchor pane for a corner direction within
// windowTarget (empty = current window), excluding excludePaneID. It returns the
// anchor pane id and the vertical join direction (top/bottom) to use against it.
// anchor is "" when no candidate pane exists (e.g. the source is the only pane).
func cornerAnchorPane(windowTarget, direction, excludePaneID string) (anchor, joinDir string, err error) {
	spec, ok := corners[direction]
	if !ok {
		return "", "", fmt.Errorf("invalid direction: %s", direction)
	}

	lines, err := tmux.ListPanes(&tmux.ListPanesParams{
		Target: windowTarget,
		Format: "#{pane_id}\t#{pane_at_left}\t#{pane_at_right}\t#{pane_top}",
	})
	if err != nil {
		return "", "", err
	}

	anchorTop := 0
	for _, line := range lines {
		fields := strings.Split(line, "\t")
		if len(fields) < 4 {
			continue
		}
		id, atLeft, atRight, topStr := fields[0], fields[1], fields[2], fields[3]
		if id == excludePaneID {
			continue
		}

		onEdge := atLeft == "1"
		if spec.atRight {
			onEdge = atRight == "1"
		}
		if !onEdge {
			continue
		}

		top, err := strconv.Atoi(topStr)
		if err != nil {
			continue
		}

		if anchor == "" || (spec.bottom && top > anchorTop) || (!spec.bottom && top < anchorTop) {
			anchor = id
			anchorTop = top
		}
	}

	joinDir = "top"
	if spec.bottom {
		joinDir = "bottom"
	}
	return anchor, joinDir, nil
}

// movePaneToCorner joins srcPane into the corner of windowTarget (empty = current
// window) named by direction, stacking it above/below the corner-most edge pane.
// It is a no-op when no anchor exists (srcPane already fills the window).
func movePaneToCorner(srcPane, windowTarget, direction, size string) error {
	anchor, joinDir, err := cornerAnchorPane(windowTarget, direction, srcPane)
	if err != nil {
		return err
	}
	if anchor == "" {
		return nil
	}
	return joinPaneInDirection(srcPane, anchor, joinDir, size, false)
}

// showPane restores a hidden pane (by its "<window_index>.<pane_index>" ref) into location,
// split according to direction (bottom/top/right/left or a corner) and an optional size.
func showPane(location, ref, direction, size string) error {
	src := hidden_session_name + ":" + ref
	if isCornerDirection(direction) {
		count, err := windowPaneCount(location)
		if err != nil {
			return err
		}
		if count <= 1 {
			// A single destination pane can't form a real corner from one join;
			// honor the corner's horizontal side instead of stacking full-width.
			return joinPaneInDirection(src, location, cornerHorizontalDirection(direction), size, false)
		}
		return movePaneToCorner(src, location, direction, size)
	}
	return joinPaneInDirection(src, location, direction, size, false)
}

// windowPaneCount returns the number of panes in windowTarget ("" = current window).
func windowPaneCount(windowTarget string) (int, error) {
	lines, err := tmux.ListPanes(&tmux.ListPanesParams{
		Target: windowTarget,
		Format: "#{pane_id}",
	})
	if err != nil {
		return 0, err
	}
	return len(lines), nil
}

// anyOtherPane returns the id of the first pane in the current window that is not
// excludePaneID, or "" when the excluded pane is the only one in the window.
func anyOtherPane(excludePaneID string) (string, error) {
	lines, err := tmux.ListPanes(&tmux.ListPanesParams{
		Format: "#{pane_id}",
	})
	if err != nil {
		return "", err
	}

	for _, id := range lines {
		if id != excludePaneID {
			return id, nil
		}
	}

	return "", nil
}

// movePaneRelative moves srcPaneID next to a target pane in the given direction.
// When target is set, it joins directly. Otherwise it shows the display-panes
// picker, recording the chosen pane into a global option, then joins in-process.
func movePaneRelative(srcPaneID, target, direction, size string) error {
	if target != "" {
		return joinPaneInDirection(srcPaneID, target, direction, size, false)
	}

	// Validate direction up front so an invalid value fails before the picker.
	if _, _, _, err := paneSplitFlags(direction); err != nil {
		return err
	}

	if err := tmux.SetGlobalOption(paneMoveTargetOption, ""); err != nil {
		return err
	}
	defer tmux.UnsetGlobalOption(paneMoveTargetOption)

	if err := tmux.DisplayPanes(&tmux.DisplayPanesParams{
		Duration: "0",
		Template: "set-option -g " + paneMoveTargetOption + " '%%'",
	}); err != nil {
		return err
	}

	chosenTarget, err := tmux.DisplayMessage("#{"+paneMoveTargetOption+"}", &tmux.DisplayMessageParams{})
	if err != nil {
		return err
	}
	if chosenTarget == "" {
		// Picker cancelled or timed out.
		return nil
	}

	return joinPaneInDirection(srcPaneID, chosenTarget, direction, size, false)
}

// movePaneToEdge moves srcPaneID to the window edge in the given direction so it
// spans the entire edge (full width for top/bottom, full height for left/right),
// using join-pane's -f (full-window span) flag. A corner direction instead docks
// the pane into that corner (within the corner-most pane's column, not full span).
func movePaneToEdge(srcPaneID, direction, size string) error {
	if isCornerDirection(direction) {
		return movePaneToCorner(srcPaneID, "", direction, size)
	}

	target, err := anyOtherPane(srcPaneID)
	if err != nil {
		return err
	}
	if target == "" {
		// Source is the only pane; it already fills the window.
		return nil
	}

	return joinPaneInDirection(srcPaneID, target, direction, size, true)
}
