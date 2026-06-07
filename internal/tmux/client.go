package tmux

type SwitchClientParams struct {
	TargetSession string
}

// SwitchClient runs `switch-client`.
func SwitchClient(params *SwitchClientParams) error {
	args := []string{"switch-client"}
	if params.TargetSession != "" {
		args = append(args, "-t", params.TargetSession)
	}

	return run(args...)
}
