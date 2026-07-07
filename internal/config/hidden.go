package config

import (
	"fmt"
	"regexp"
)

// HiddenSessionNamePatternKey is the config key for the hidden-session regexp.
const HiddenSessionNamePatternKey = "session.hidden_name_pattern"

const defaultHiddenSessionNamePattern = `^_(.*_)?$`

// HiddenSessionMatcher returns a predicate reporting whether a session is hidden.
// HiddenSessionName is always hidden; an invalid configured pattern errors.
func HiddenSessionMatcher() (func(name string) bool, error) {
	pattern := Get[string](HiddenSessionNamePatternKey)
	if pattern == "" {
		pattern = defaultHiddenSessionNamePattern
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid %s %q: %w", HiddenSessionNamePatternKey, pattern, err)
	}

	return func(name string) bool {
		return name == HiddenSessionName || re.MatchString(name)
	}, nil
}
