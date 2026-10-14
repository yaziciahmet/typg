package query

import "slices"

// Table identifies a database table. Codegen implements this for each table.
type Table interface {
	SQLName() string
	Alias() string
}

// Column identifies a table column or a column-like expression. Codegen
// implements it for each table column.
type Column interface {
	Qualifier() string
	SQLName() string
}

type qualColumn struct {
	table  string
	column string
}

// Except removes the columns in drop from cols in place, matched by qualified
// name, and returns the shortened slice. Do not use cols afterwards.
func Except(cols []Column, drop ...Column) []Column {
	return slices.DeleteFunc(cols, func(c Column) bool {
		q := c.Qualifier()
		return slices.ContainsFunc(drop, func(d Column) bool { return d.Qualifier() == q })
	})
}

// Col references a qualified column outside generated schema (e.g. CTE columns).
func Col(table, column string) Column {
	return qualColumn{table: table, column: column}
}

func (q qualColumn) SQLName() string   { return q.column }
func (q qualColumn) Qualifier() string { return q.table + "." + q.column }

// RawColumn is a SQL expression with ? placeholders replaced by numbered args
// at compile time, usable wherever a column is: select list, WHERE, HAVING,
// GROUP BY, ORDER BY. Write ?? for a literal question mark. Use As to name it
// in the select list; ORDER BY then refers to the alias.
type RawColumn struct {
	sql   string
	args  []any
	alias string
}

// NewRawCol builds a raw SQL expression. Prefer As over writing AS in sql.
func NewRawCol(sql string, args ...any) *RawColumn {
	return &RawColumn{sql: sql, args: args}
}

// As returns a copy that renders as "expr AS alias" in the select list.
func (r *RawColumn) As(alias string) *RawColumn {
	c := *r
	c.alias = alias
	return &c
}

func (r *RawColumn) SQLName() string   { return r.sql }
func (r *RawColumn) Qualifier() string { return r.sql }

func (r *RawColumn) compileExpr(c *Compiler) {
	writeInterpolated(c, r.sql, r.args)
}

func (r *RawColumn) compileSelect(c *Compiler) {
	r.compileExpr(c)
	if r.alias != "" {
		c.Write(" AS ")
		c.Write(r.alias)
	}
}

func (r *RawColumn) aliasName() string { return r.alias }

type aliasedColumn struct {
	col   Column
	alias string
}

// ColumnAs renders col as "col AS alias" in the select list. Elsewhere it
// behaves as col, except ORDER BY, which refers to the alias.
func ColumnAs(col Column, alias string) Column {
	if r, ok := col.(*RawColumn); ok {
		return r.As(alias)
	}
	return aliasedColumn{col: col, alias: alias}
}

func (a aliasedColumn) SQLName() string   { return a.col.SQLName() }
func (a aliasedColumn) Qualifier() string { return a.col.Qualifier() }

func (a aliasedColumn) compileExpr(c *Compiler) { writeExpr(c, a.col) }

func (a aliasedColumn) compileSelect(c *Compiler) {
	writeExpr(c, a.col)
	c.Write(" AS ")
	c.Write(a.alias)
}

func (a aliasedColumn) aliasName() string { return a.alias }

type rawTable string

func (r rawTable) SQLName() string { return string(r) }
func (r rawTable) Alias() string   { return string(r) }

// Named returns a table reference by name (CTEs, ad-hoc).
func Named(name string) Table { return rawTable(name) }
