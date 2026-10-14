package query

type andExpr struct {
	exprs []Expr
}

// And joins expressions with AND. With no expressions it renders TRUE, the
// identity of AND.
func And(exprs ...Expr) Expr {
	switch len(exprs) {
	case 0:
		return boolExpr(true)
	case 1:
		return exprs[0]
	}
	return andExpr{exprs: exprs}
}

func (a andExpr) compile(c *Compiler) {
	for i, e := range a.exprs {
		if i > 0 {
			c.Write(" AND ")
		}
		c.Write("(")
		e.compile(c)
		c.Write(")")
	}
}

type orExpr struct {
	exprs []Expr
}

// Or joins expressions with OR. With no expressions it renders FALSE, the
// identity of OR.
func Or(exprs ...Expr) Expr {
	switch len(exprs) {
	case 0:
		return boolExpr(false)
	case 1:
		return exprs[0]
	}
	return orExpr{exprs: exprs}
}

type boolExpr bool

func (b boolExpr) compile(c *Compiler) {
	if b {
		c.Write("TRUE")
		return
	}
	c.Write("FALSE")
}

func (o orExpr) compile(c *Compiler) {
	for i, e := range o.exprs {
		if i > 0 {
			c.Write(" OR ")
		}
		c.Write("(")
		e.compile(c)
		c.Write(")")
	}
}

type inExpr struct {
	col  Column
	vals []any
	neg  bool
}

// In builds col IN ($1, $2, ...). With no values it renders FALSE.
func In(col Column, values ...any) Expr {
	return inExpr{col: col, vals: values}
}

// NotIn builds col NOT IN ($1, $2, ...). With no values it renders TRUE.
func NotIn(col Column, values ...any) Expr {
	return inExpr{col: col, vals: values, neg: true}
}

func (e inExpr) compile(c *Compiler) {
	if len(e.vals) == 0 {
		if e.neg {
			c.Write("TRUE")
		} else {
			c.Write("FALSE")
		}
		return
	}
	writeExpr(c, e.col)
	if e.neg {
		c.Write(" NOT IN (")
	} else {
		c.Write(" IN (")
	}
	for i, v := range e.vals {
		if i > 0 {
			c.Write(", ")
		}
		c.Write(c.Arg(v))
	}
	c.Write(")")
}

type likeExpr struct {
	col     Column
	pattern string
	ilike   bool
}

// Like builds col LIKE pattern.
func Like(col Column, pattern string) Expr {
	return likeExpr{col: col, pattern: pattern}
}

// ILike builds col ILIKE pattern (postgres).
func ILike(col Column, pattern string) Expr {
	return likeExpr{col: col, pattern: pattern, ilike: true}
}

func (e likeExpr) compile(c *Compiler) {
	writeExpr(c, e.col)
	if e.ilike {
		c.Write(" ILIKE ")
	} else {
		c.Write(" LIKE ")
	}
	c.Write(c.Arg(e.pattern))
}

type compareExpr struct {
	left     Column
	op       string
	rightCol Column
	rightVal any
	isCol    bool
}

func compareVal(col Column, op string, val any) Expr {
	return compareExpr{left: col, op: op, rightVal: val}
}

func compareCol(col Column, op string, other Column) Expr {
	return compareExpr{left: col, op: op, rightCol: other, isCol: true}
}

// Eq builds col = val.
func Eq(col Column, val any) Expr { return compareVal(col, "=", val) }

// EqCol builds col = other.
func EqCol(col, other Column) Expr { return compareCol(col, "=", other) }

// Ne builds col != val.
func Ne(col Column, val any) Expr { return compareVal(col, "!=", val) }

// NeCol builds col != other.
func NeCol(col, other Column) Expr { return compareCol(col, "!=", other) }

// Gt builds col > val.
func Gt(col Column, val any) Expr { return compareVal(col, ">", val) }

// GtCol builds col > other.
func GtCol(col, other Column) Expr { return compareCol(col, ">", other) }

// Gte builds col >= val.
func Gte(col Column, val any) Expr { return compareVal(col, ">=", val) }

// GteCol builds col >= other.
func GteCol(col, other Column) Expr { return compareCol(col, ">=", other) }

// Lt builds col < val.
func Lt(col Column, val any) Expr { return compareVal(col, "<", val) }

// LtCol builds col < other.
func LtCol(col, other Column) Expr { return compareCol(col, "<", other) }

// Lte builds col <= val.
func Lte(col Column, val any) Expr { return compareVal(col, "<=", val) }

// LteCol builds col <= other.
func LteCol(col, other Column) Expr { return compareCol(col, "<=", other) }

func (e compareExpr) compile(c *Compiler) {
	writeExpr(c, e.left)
	c.Write(" ")
	c.Write(e.op)
	c.Write(" ")
	if e.isCol {
		writeExpr(c, e.rightCol)
	} else {
		c.Write(c.Arg(e.rightVal))
	}
}

type nullExpr struct {
	col Column
	neg bool
}

// IsNull builds col IS NULL.
func IsNull(col Column) Expr { return nullExpr{col: col} }

// NotNull builds col IS NOT NULL.
func NotNull(col Column) Expr { return nullExpr{col: col, neg: true} }

func (e nullExpr) compile(c *Compiler) {
	writeExpr(c, e.col)
	if e.neg {
		c.Write(" IS NOT NULL")
	} else {
		c.Write(" IS NULL")
	}
}

type containsExpr struct {
	col Column
	val any
}

// Contains builds col @> ARRAY[val] (postgres array containment).
func Contains(col Column, val any) Expr {
	return containsExpr{col: col, val: val}
}

func (e containsExpr) compile(c *Compiler) {
	writeExpr(c, e.col)
	c.Write(" @> ARRAY[")
	c.Write(c.Arg(e.val))
	c.Write("]")
}

type notExpr struct {
	expr Expr
}

// Not builds NOT (expr).
func Not(e Expr) Expr { return notExpr{expr: e} }

func (n notExpr) compile(c *Compiler) {
	c.Write("NOT (")
	n.expr.compile(c)
	c.Write(")")
}

type betweenExpr struct {
	col    Column
	lo, hi any
}

// Between builds col BETWEEN lo AND hi.
func Between(col Column, lo, hi any) Expr {
	return betweenExpr{col: col, lo: lo, hi: hi}
}

func (e betweenExpr) compile(c *Compiler) {
	writeExpr(c, e.col)
	c.Write(" BETWEEN ")
	writeArg(c, e.lo)
	c.Write(" AND ")
	writeArg(c, e.hi)
}

type eqAnyExpr struct {
	col    Column
	values any
}

// EqAny builds col = ANY($1), binding values (a slice) as one array argument.
func EqAny(col Column, values any) Expr {
	return eqAnyExpr{col: col, values: values}
}

func (e eqAnyExpr) compile(c *Compiler) {
	writeExpr(c, e.col)
	c.Write(" = ANY(")
	c.Write(c.Arg(e.values))
	c.Write(")")
}

type existsExpr struct {
	sub *SelectBuilder
}

// Exists builds EXISTS (subquery).
func Exists(sub *SelectBuilder) Expr { return existsExpr{sub: sub} }

func (e existsExpr) compile(c *Compiler) {
	c.Write("EXISTS (")
	e.sub.compile(c)
	c.Write(")")
}

type inSelectExpr struct {
	col Column
	sub *SelectBuilder
}

// InSelect builds col IN (subquery).
func InSelect(col Column, sub *SelectBuilder) Expr {
	return inSelectExpr{col: col, sub: sub}
}

func (e inSelectExpr) compile(c *Compiler) {
	writeExpr(c, e.col)
	c.Write(" IN (")
	e.sub.compile(c)
	c.Write(")")
}
