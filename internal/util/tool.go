package util

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var missingToolHelpText = map[string]string{
	"git": strings.TrimSpace(`
Homebrew:
	brew install git
`),
	"fzf": strings.TrimSpace(`
Homebrew:
  brew install fzf
`),
	"gh": strings.TrimSpace(`
Homebrew:
	brew install gh

Or see: https://cli.github.com
`),
	"tmux": strings.TrimSpace(`
Homebrew:
	brew install tmux
`),
}

// HasTool reports whether tool is available on $PATH.
func HasTool(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}

func EnsureTool(tool string) error {
	_, err := exec.LookPath(tool)
	if err == nil {
		return nil
	}

	var buf bytes.Buffer
	buf.WriteString(err.Error())
	buf.WriteString("\n\n")

	fmt.Fprintf(&buf, "Missing tool: '%s'.\n", tool)
	if helpText, ok := missingToolHelpText[tool]; ok {
		fmt.Fprintf(&buf, "\n%s\n", helpText)
	}
	return errors.New(buf.String())
}
