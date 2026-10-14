package query

func funcCol(sql string, args ...any) *RawColumn {
	return &RawColumn{sql: sql, args: args, atomic: true}
}

// Count builds count(*).
func Count() *RawColumn { return funcCol("count(*)") }

// CountOf builds count(col).
func CountOf(col Column) *RawColumn { return funcCol("count(?)", col) }

// Sum builds sum(col).
func Sum(col Column) *RawColumn { return funcCol("sum(?)", col) }

// Max builds max(col).
func Max(col Column) *RawColumn { return funcCol("max(?)", col) }

// Min builds min(col).
func Min(col Column) *RawColumn { return funcCol("min(?)", col) }

// Coalesce builds COALESCE(col, fallback); fallback is a bound value, or a column.
func Coalesce(col Column, fallback any) *RawColumn {
	return funcCol("COALESCE(?, ?)", col, fallback)
}
