package schema

import (
	"strings"
	"unicode"
)

// ExportIdent turns a SQL name into an exported Go identifier: letter and digit
// runs become capitalized words, id/url/uuid become ID/URL/UUID, and a leading
// digit gets an X prefix.
func ExportIdent(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
	for i, w := range words {
		switch strings.ToLower(w) {
		case "id":
			words[i] = "ID"
		case "url":
			words[i] = "URL"
		case "uuid":
			words[i] = "UUID"
		default:
			runes := []rune(w)
			runes[0] = unicode.ToUpper(runes[0])
			words[i] = string(runes)
		}
	}
	out := strings.Join(words, "")
	if out == "" || unicode.IsDigit([]rune(out)[0]) {
		out = "X" + out
	}
	return out
}
