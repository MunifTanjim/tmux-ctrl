package tmux

type ListWindowsParams struct {
	Format        string
	TargetSession string
}

// ListWindows runs `list-windows` and returns the lines.
func ListWindows(params *ListWindowsParams) ([]string, error) {
	args := []string{"list-windows"}
	if params.Format != "" {
		args = append(args, "-F", params.Format)
	}
	if params.TargetSession != "" {
		args = append(args, "-t", params.TargetSession)
	}

	return queryLines(args...)
}
