package tmux

type HasSessionParams struct {
	TargetSession string
}

// HasSession reports whether the target session exists.
func HasSession(params *HasSessionParams) bool {
	return run("has-session", "-t", params.TargetSession) == nil
}

type NewSessionParams struct {
	SessionName string
	Detached    bool
}

// NewSession runs `new-session`.
func NewSession(params *NewSessionParams) error {
	args := []string{"new-session"}
	if params.Detached {
		args = append(args, "-d")
	}
	if params.SessionName != "" {
		args = append(args, "-s", params.SessionName)
	}

	return run(args...)
}

type ListSessionsParams struct {
	Format string
}

// ListSessions runs `list-sessions` and returns the lines.
func ListSessions(params *ListSessionsParams) ([]string, error) {
	args := []string{"list-sessions"}
	if params.Format != "" {
		args = append(args, "-F", params.Format)
	}

	return queryLines(args...)
}
