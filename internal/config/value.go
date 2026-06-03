package config

import (
	"path/filepath"
	"strings"
)

func ExpandHome(path string) string {
	if path == "" || HomeDir == "" {
		return path
	}
	if path[0] == '~' {
		return filepath.Join(HomeDir, path[1:])
	}
	if strings.HasPrefix(path, "$HOME") {
		return filepath.Join(HomeDir, path[5:])
	}
	if strings.HasPrefix(path, "${HOME}") {
		return filepath.Join(HomeDir, path[7:])
	}
	return path
}
