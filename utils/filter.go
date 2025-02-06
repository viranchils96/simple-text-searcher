package utils

import (
	"strings"
)

var (
	stopwords = initStopwords()
)

func initStopwords() map[string]struct{} {
	// Load from file or embed
	return map[string]struct{}{
		"a": {}, "and": {}, "be": {}, "have": {}, "i": {},
		"in": {}, "of": {}, "that": {}, "the": {}, "to": {},
	}
}

// Memory-efficient stopword filter
func stopwordFilter(tokens []string) []string {
	n := 0
	for _, token := range tokens {
		if _, ok := stopwords[token]; !ok {
			tokens[n] = token
			n++
		}
	}
	return tokens[:n]
}

// In-place lowercase transformation
func lowercaseFilter(tokens []string) []string {
	for i := range tokens {
		tokens[i] = strings.ToLower(tokens[i])
	}
	return tokens
}
