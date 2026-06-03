package util

import "strings"

func getOrCreateMap(parent map[string]any, key string) map[string]any {
	if child, ok := parent[key].(map[string]any); ok {
		return child
	}
	child := make(map[string]any)
	parent[key] = child
	return child
}

func SetNestedMapValue(m map[string]any, key string, value string) {
	parts := strings.Split(key, ".")
	parent := m
	for _, part := range parts[:len(parts)-1] {
		parent = getOrCreateMap(parent, part)
	}
	parent[parts[len(parts)-1]] = value
}

func GetNestedMapValue(m map[string]any, key string) string {
	parts := strings.Split(key, ".")
	current := m
	for _, part := range parts[:len(parts)-1] {
		if child, ok := current[part].(map[string]any); ok {
			current = child
		} else {
			return ""
		}
	}
	if value, ok := current[parts[len(parts)-1]].(string); ok {
		return value
	}
	return ""
}
