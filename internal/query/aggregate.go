package query

// Count builds count(*).
func Count() *RawColumn { return NewRawCol("count(*)") }

// CountOf builds count(col).
func CountOf(col Column) *RawColumn { return NewRawCol("count(?)", col) }

// Sum builds sum(col).
func Sum(col Column) *RawColumn { return NewRawCol("sum(?)", col) }

// Max builds max(col).
func Max(col Column) *RawColumn { return NewRawCol("max(?)", col) }

// Min builds min(col).
func Min(col Column) *RawColumn { return NewRawCol("min(?)", col) }

// Coalesce builds COALESCE(col, fallback); fallback is a bound value, or a column.
func Coalesce(col Column, fallback any) *RawColumn {
	return NewRawCol("COALESCE(?, ?)", col, fallback)
}
