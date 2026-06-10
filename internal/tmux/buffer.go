package tmux

// SetBuffer runs `set-buffer` to put value into the top tmux paste buffer.
func SetBuffer(value string) error {
	return run("set-buffer", "--", value)
}
