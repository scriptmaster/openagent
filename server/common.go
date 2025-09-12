package server

import (
	"unicode"
)

// Title capitalizes the first letter of each word in a string
// This replaces the deprecated strings.Title function
func Title(s string) string {
	if s == "" {
		return s
	}

	// Convert to runes to handle Unicode properly
	runes := []rune(s)

	// Capitalize first letter
	if len(runes) > 0 {
		runes[0] = unicode.ToUpper(runes[0])
	}

	// Capitalize letters after spaces
	for i := 1; i < len(runes); i++ {
		if unicode.IsSpace(runes[i-1]) {
			runes[i] = unicode.ToUpper(runes[i])
		}
	}

	return string(runes)
}
