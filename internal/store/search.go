package store

import "strings"

func searchTerms(query string) []string {
	raw := strings.Fields(strings.TrimSpace(query))
	if len(raw) == 0 {
		return nil
	}
	return raw
}

func likeContainsPattern(term string) string {
	term = strings.ReplaceAll(term, `\`, `\\`)
	term = strings.ReplaceAll(term, `%`, `\%`)
	term = strings.ReplaceAll(term, `_`, `\_`)
	return "%" + term + "%"
}
