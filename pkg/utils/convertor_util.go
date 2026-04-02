package utils

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

func GenerateSlug(title string) string {
	slug := strings.ToLower(title)
	var b strings.Builder
	for _, r := range slug {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else if r == ' ' || r == '-' {
			b.WriteRune('-')
		}
	}
	result := b.String()
	// Remove consecutive dashes
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}
	result = strings.Trim(result, "-")
	// Append timestamp suffix for uniqueness
	return fmt.Sprintf("%s-%d", result, time.Now().UnixMilli()%100000)
}
