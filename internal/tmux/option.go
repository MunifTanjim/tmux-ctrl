package tmux

// SetGlobalOption runs `set-option -g` to set a global option.
func SetGlobalOption(option, value string) error {
	return run("set-option", "-g", option, value)
}

// UnsetGlobalOption runs `set-option -gu` to unset a global option.
func UnsetGlobalOption(option string) error {
	return run("set-option", "-gu", option)
}
