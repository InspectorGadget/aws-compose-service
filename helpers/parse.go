package helpers

import "strings"

// WithFallbackValue returns value if non-empty, otherwise fallback.
func WithFallbackValue(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

// SplitAndTrim splits a comma-separated string and trims spaces, dropping empties.
func SplitAndTrim(input string) []string {
	if input == "" {
		return nil
	}

	parts := strings.Split(input, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
