package common

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

func GenerateSlug(s string) string {
	s = strings.ToLower(s)

	s = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || r == '_' {
			return '-'
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			return r
		}
		return -1
	}, s)

	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}

	s = strings.Trim(s, "-")

	timestamp := time.Now().UnixNano() / int64(time.Millisecond)

	return fmt.Sprintf("%s-%d", s, timestamp)
}