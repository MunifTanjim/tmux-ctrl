package tmux

type DisplayMessageParams struct {
	TargetPane string
}

// DisplayMessage runs `display-message -p` for the given format and returns the result.
// An empty target omits the `-t` flag (defaulting to the current pane).
func DisplayMessage(format string, params *DisplayMessageParams) (string, error) {
	args := []string{"display-message", "-p"}
	if params.TargetPane != "" {
		args = append(args, "-t", params.TargetPane)
	}
	args = append(args, format)

	return query(args...)
}
