package query

import (
	"strconv"
	"strings"
)

// fragment is anything that renders SQL through a Compiler.
type fragment interface {
	compile(c *Compiler)
}

// Expr is a boolean SQL expression for WHERE, ON, HAVING, etc.
type Expr interface {
	fragment
}

// Compiler renders SQL and collects placeholder arguments.
type Compiler struct {
	buf   strings.Builder
	args  []any
	index int
}

// NewCompiler returns a compiler for PostgreSQL ($1, $2, ...).
func NewCompiler() *Compiler {
	return &Compiler{}
}

// Write appends literal SQL.
func (c *Compiler) Write(sql string) {
	c.buf.WriteString(sql)
}

// Arg registers a value and returns its PostgreSQL placeholder.
func (c *Compiler) Arg(v any) string {
	c.index++
	c.args = append(c.args, v)
	return "$" + strconv.Itoa(c.index)
}

// SQL returns the rendered query string.
func (c *Compiler) SQL() string {
	return c.buf.String()
}

// Args returns collected arguments in placeholder order.
func (c *Compiler) Args() []any {
	return c.args
}

func writeArg(c *Compiler, v any) {
	if col, ok := v.(Column); ok {
		c.Write(col.Qualifier())
		return
	}
	c.Write(c.Arg(v))
}

func writeInterpolated(c *Compiler, sql string, args []any) {
	if len(args) == 0 && !strings.Contains(sql, "??") {
		c.Write(sql)
		return
	}

	argIdx := 0
	start := 0
	for i := 0; i < len(sql); i++ {
		if sql[i] != '?' {
			continue
		}
		if i+1 < len(sql) && sql[i+1] == '?' {
			c.Write(sql[start : i+1])
			start = i + 2
			i++
			continue
		}
		if argIdx < len(args) {
			if i > start {
				c.Write(sql[start:i])
			}
			writeArg(c, args[argIdx])
			argIdx++
			start = i + 1
		}
	}
	if start < len(sql) {
		c.Write(sql[start:])
	}
}

// Raw is a SQL fragment with ? placeholders replaced by numbered args.
// Write ?? for a literal question mark.
type Raw struct {
	sql  string
	args []any
}

// RawExpr builds an expression from a SQL template and arguments.
func RawExpr(sql string, args ...any) Expr {
	return Raw{sql: sql, args: args}
}

func (r Raw) compile(c *Compiler) {
	writeInterpolated(c, r.sql, r.args)
}

type columnFrag struct {
	col Column
}

func (f columnFrag) compile(c *Compiler) {
	if cc, ok := f.col.(interface{ compile(c *Compiler) }); ok {
		cc.compile(c)
		return
	}
	c.Write(f.col.Qualifier())
}

type tableFrag struct {
	table Table
}

func (f tableFrag) compile(c *Compiler) {
	c.Write(f.table.SQLName())
	if alias := f.table.Alias(); alias != "" && alias != f.table.SQLName() {
		c.Write(" AS ")
		c.Write(alias)
	}
}

// Assignment sets a column to a bound value, a SQL expression, or EXCLUDED.col.
// Generated columns expose typed constructors: Set, SetNull, SetExpr, SetExcluded.
type Assignment struct {
	name     string
	val      any
	expr     string
	args     []any
	excluded bool
}

// Assign sets col to val.
func Assign(col Column, val any) Assignment {
	return Assignment{name: col.SQLName(), val: val}
}

// AssignExpr sets col to a SQL expression, with ? placeholders bound to args.
func AssignExpr(col Column, sql string, args ...any) Assignment {
	return Assignment{name: col.SQLName(), expr: sql, args: args}
}

// AssignExcluded sets col to EXCLUDED.col, for ON CONFLICT DO UPDATE.
func AssignExcluded(col Column) Assignment {
	return Assignment{name: col.SQLName(), excluded: true}
}

func (a Assignment) compileRHS(c *Compiler) {
	switch {
	case a.excluded:
		c.Write("EXCLUDED.")
		c.Write(a.name)
	case a.expr != "":
		writeInterpolated(c, a.expr, a.args)
	default:
		c.Write(c.Arg(a.val))
	}
}

func mergeAssignments(sets []Assignment, assigns ...Assignment) []Assignment {
	for _, a := range assigns {
		replaced := false
		for i, s := range sets {
			if s.name == a.name {
				sets[i] = a
				replaced = true
				break
			}
		}
		if !replaced {
			sets = append(sets, a)
		}
	}
	return sets
}

func compilePrefix(c *Compiler, with *WithBuilder, prefix []fragment) {
	if with != nil {
		with.compilePrefix(c)
	}
	for _, p := range prefix {
		p.compile(c)
	}
}

func compileSuffix(c *Compiler, suffix []fragment) {
	for _, p := range suffix {
		p.compile(c)
	}
}

func compileWhere(c *Compiler, wheres []Expr) {
	if len(wheres) == 0 {
		return
	}
	c.Write(" WHERE ")
	compileConjunction(c, wheres)
}

func compileHaving(c *Compiler, having []Expr) {
	if len(having) == 0 {
		return
	}
	c.Write(" HAVING ")
	compileConjunction(c, having)
}

func compileConjunction(c *Compiler, exprs []Expr) {
	for i, e := range exprs {
		if i > 0 {
			c.Write(" AND ")
		}
		if len(exprs) > 1 && needsParens(e) {
			c.Write("(")
			e.compile(c)
			c.Write(")")
			continue
		}
		e.compile(c)
	}
}

func needsParens(e Expr) bool {
	switch e.(type) {
	case orExpr, Raw:
		return true
	}
	return false
}

func compileReturning(c *Compiler, returns []fragment) {
	if len(returns) == 0 {
		return
	}
	c.Write(" RETURNING ")
	for i, col := range returns {
		if i > 0 {
			c.Write(", ")
		}
		col.compile(c)
	}
}

func build(c *Compiler, stmt fragment) (string, []any) {
	stmt.compile(c)
	return c.SQL(), c.Args()
}
