package codegen

import (
	"strings"
	"unicode"

	"github.com/yaziciahmet/typg/internal/schema"
)

func tableVar(name string) string {
	return schema.ExportIdent(name)
}

func tableStruct(name string) string {
	return lowerFirst(tableVar(name)) + "Table"
}

var tableMembers = map[string]bool{
	"All": true, "AllExcept": true, "Alias": true, "As": true,
	"Insert": true, "Update": true, "Delete": true, "SQLName": true,
}

func columnField(name string) string {
	field := schema.ExportIdent(name)
	if tableMembers[field] {
		field += "_"
	}
	return field
}

func columnType(table, column string) string {
	return lowerFirst(tableVar(table)) + columnField(column) + "Col"
}

func modelName(table string) string {
	return tableVar(table)
}

func enumConstName(typeName, label string) string {
	return typeName + schema.ExportIdent(label)
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// quoteIdent quotes a SQL identifier only when it has to be: mixed case,
// special characters, or a reserved word.
func quoteIdent(name string) string {
	if isPlainIdent(name) && !reservedWords[name] {
		return name
	}
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func isPlainIdent(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		switch {
		case r == '_' || (r >= 'a' && r <= 'z'):
		case i > 0 && r >= '0' && r <= '9':
		default:
			return false
		}
	}
	return true
}

var reservedWords = func() map[string]bool {
	m := map[string]bool{}
	for w := range strings.FieldsSeq(`all analyse analyze and any array as asc asymmetric authorization binary both case cast
		check collate collation column concurrently constraint create cross current_catalog current_date current_role
		current_schema current_time current_timestamp current_user default deferrable desc distinct do else end except
		false fetch for foreign freeze from full grant group having ilike in initially inner intersect into is isnull
		join lateral leading left like limit localtime localtimestamp natural not notnull null offset on only or order
		outer overlaps placing primary references returning right select session_user similar some symmetric system_user
		table tablesample then to trailing true union unique user using variadic verbose when where window with`) {
		m[w] = true
	}
	return m
}()
