package query

import "testing"

func TestCompilerPlaceholderNumbering(t *testing.T) {
	c := NewCompiler()
	c.Write("SELECT * FROM users WHERE ")
	RawExpr("id = ?", 1).compile(c)
	c.Write(" AND ")
	RawExpr("name = ?", "alice").compile(c)

	if got := c.SQL(); got != "SELECT * FROM users WHERE id = $1 AND name = $2" {
		t.Fatalf("sql = %q", got)
	}
	if args := c.Args(); len(args) != 2 || args[0] != 1 || args[1] != "alice" {
		t.Fatalf("args = %v", args)
	}
}

func TestCompilerManyPlaceholders(t *testing.T) {
	c := NewCompiler()
	args := make([]any, 12)
	for i := range args {
		args[i] = i
	}
	RawExpr("? ? ? ? ? ? ? ? ? ? ? ?", args...).compile(c)

	want := "$1 $2 $3 $4 $5 $6 $7 $8 $9 $10 $11 $12"
	if got := c.SQL(); got != want {
		t.Fatalf("sql = %q", got)
	}
	if len(c.Args()) != 12 {
		t.Fatalf("args len = %d", len(c.Args()))
	}
}

func TestWriteArgColumn(t *testing.T) {
	c := NewCompiler()
	writeInterpolated(c, "COALESCE(?, ?)", []any{Col("users", "verified_name"), Col("users", "provided_name")})

	want := "COALESCE(users.verified_name, users.provided_name)"
	if got := c.SQL(); got != want {
		t.Fatalf("sql = %q", got)
	}
	if len(c.Args()) != 0 {
		t.Fatalf("args = %v", c.Args())
	}
}

func TestWriteInterpolatedNoArgs(t *testing.T) {
	c := NewCompiler()
	writeInterpolated(c, "count(*) OVER (PARTITION BY status)", nil)

	want := "count(*) OVER (PARTITION BY status)"
	if got := c.SQL(); got != want {
		t.Fatalf("sql = %q", got)
	}
	if len(c.Args()) != 0 {
		t.Fatalf("args = %v", c.Args())
	}
}

func TestRawExprLiteralQuestionMark(t *testing.T) {
	c := NewCompiler()
	RawExpr("jsonb ? key", "??").compile(c)

	if got := c.SQL(); got != "jsonb $1 key" {
		t.Fatalf("sql = %q", got)
	}
	if args := c.Args(); len(args) != 1 || args[0] != "??" {
		t.Fatalf("args = %v", args)
	}
}

func TestWriteInterpolatedEscapedQuestionMark(t *testing.T) {
	cases := []struct {
		name string
		sql  string
		args []any
		want string
	}{
		{"with args", "meta ?? ? AND x = ?", []any{"k", 1}, "meta ? $1 AND x = $2"},
		{"without args", "meta ?? 'k'", nil, "meta ? 'k'"},
		{"single question mark without args stays literal", "meta ? 'k'", nil, "meta ? 'k'"},
		{"single question mark after args are used up stays literal", "x = ? AND meta ? 'k'", []any{1}, "x = $1 AND meta ? 'k'"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := NewCompiler()
			writeInterpolated(c, tc.sql, tc.args)
			if got := c.SQL(); got != tc.want {
				t.Fatalf("sql = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRawColumnTableIsNotNil(t *testing.T) {
	if NewRawCol("count(*)").Table() == nil {
		t.Fatal("Table() must not be nil")
	}
}
