package utils

import (
	"unicode"

	"github.com/kljensen/snowball"
)

func tokenize(text string) []string {
	var tokens []string
	var token []rune

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			token = append(token, r)
		} else if len(token) > 0 {
			tokens = append(tokens, string(token))
			token = token[:0]
		}
	}

	if len(token) > 0 {
		tokens = append(tokens, string(token))
	}
	return tokens
}

func analyze(text string) []string {
	tokens := tokenize(text)
	tokens = lowercaseFilter(tokens)
	tokens = stopwordFilter(tokens)
	tokens = stemmerFilter(tokens)
	return tokens
}

func stemmerFilter(tokens []string) []string {
	r := make([]string, len(tokens))
	for i, token := range tokens {
		stemmed, _ := snowball.Stem(token, "english", false)
		r[i] = stemmed
	}
	return r
}
