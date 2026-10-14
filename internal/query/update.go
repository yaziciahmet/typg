package query

// UpdateBuilder builds UPDATE statements.
type UpdateBuilder struct {
	with    *WithBuilder
	prefix  []fragment
	suffix  []fragment
	table   Table
	sets    []Assignment
	from    []Table
	wheres  []Expr
	returns []Column
}

// Update starts an UPDATE on table.
func Update(t Table) *UpdateBuilder {
	return &UpdateBuilder{table: t}
}

// Prefix injects SQL before UPDATE.
func (b *UpdateBuilder) Prefix(sql string, args ...any) *UpdateBuilder {
	b.prefix = append(b.prefix, Raw{sql: sql, args: args})
	return b
}

// Suffix injects SQL after the statement.
func (b *UpdateBuilder) Suffix(sql string, args ...any) *UpdateBuilder {
	b.suffix = append(b.suffix, Raw{sql: sql, args: args})
	return b
}

// Set adds column assignments; assigning the same column again overwrites it.
func (b *UpdateBuilder) Set(assigns ...Assignment) *UpdateBuilder {
	b.sets = mergeAssignments(b.sets, assigns...)
	return b
}

// SetAny sets a column by name.
func (b *UpdateBuilder) SetAny(column string, val any) *UpdateBuilder {
	return b.Set(Assignment{name: column, val: val})
}

// Where adds an AND condition.
func (b *UpdateBuilder) Where(e Expr) *UpdateBuilder {
	b.wheres = append(b.wheres, e)
	return b
}

// Returning adds a RETURNING clause.
func (b *UpdateBuilder) Returning(cols ...Column) *UpdateBuilder {
	b.returns = append(b.returns, cols...)
	return b
}

// From adds UPDATE ... FROM tables (postgres).
func (b *UpdateBuilder) From(tables ...Table) *UpdateBuilder {
	b.from = append(b.from, tables...)
	return b
}

// Build renders SQL and arguments.
func (b *UpdateBuilder) Build() (string, []any) {
	return build(NewCompiler(), b)
}

func (b *UpdateBuilder) compile(c *Compiler) {
	compilePrefix(c, b.with, b.prefix)
	c.Write("UPDATE ")
	tableFrag{table: b.table}.compile(c)
	c.Write(" SET ")
	for i, s := range b.sets {
		if i > 0 {
			c.Write(", ")
		}
		c.Write(s.name)
		c.Write(" = ")
		s.compileRHS(c)
	}
	if len(b.from) > 0 {
		c.Write(" FROM ")
		for i, t := range b.from {
			if i > 0 {
				c.Write(", ")
			}
			tableFrag{table: t}.compile(c)
		}
	}
	compileWhere(c, b.wheres)
	compileReturning(c, b.returns)
	compileSuffix(c, b.suffix)
}
