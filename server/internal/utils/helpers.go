package utils

import (
	"encoding/json"
	"io"
	"strings"
)

// parseJSON decodes a JSON reader into a target struct.
func parseJSON(r io.Reader, target any) error {
	return json.NewDecoder(r).Decode(target)
}

// FormatEmailToName converts "john.doe@gmail.com" → "john.doe"
// Mirrors the same helper in the original Node.js backend.
func FormatEmailToName(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) == 0 {
		return email
	}
	return parts[0]
}

// RemoveExtraSpaces collapses multiple spaces into one and trims edges.
func RemoveExtraSpaces(s string) string {
	words := strings.Fields(s) // Fields splits on any whitespace and removes empties
	return strings.Join(words, " ")
}
