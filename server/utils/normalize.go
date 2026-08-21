package utils

import "strings"

func NormalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
