package query_test

import (
	"testing"

	"github.com/yaziciahmet/typg/internal/query"
	"github.com/yaziciahmet/typg/internal/query/querytest"
)

func TestInsertSetOverwrite(t *testing.T) {
	sql, args := query.Insert(querytest.TUsers).
		Set(query.Assign(querytest.TUsersName, "first")).
		Set(query.Assign(querytest.TUsersName, "second")).
		Set(query.Assign(querytest.TUsersID, int64(1))).
		Build()

	want := "INSERT INTO users (name, id) VALUES ($1, $2)"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if args[0] != "second" || args[1] != int64(1) {
		t.Fatalf("args = %v", args)
	}
}

func TestInsertReturning(t *testing.T) {
	sql, args := query.Insert(querytest.TUsers).
		Set(query.Assign(querytest.TUsersName, "alice")).
		Returning(querytest.TUsersID).
		Build()

	want := "INSERT INTO users (name) VALUES ($1) RETURNING users.id"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 1 || args[0] != "alice" {
		t.Fatalf("args = %v", args)
	}
}

func TestInsertOnConflictDoUpdate(t *testing.T) {
	sql, args := query.Insert(querytest.TUsers).
		Set(query.Assign(querytest.TUsersEmail, "a@b.com")).
		Set(query.Assign(querytest.TUsersName, "alice")).
		OnConflict(querytest.TUsersEmail).
		DoUpdate(
			query.Assign(querytest.TUsersName, "alice"),
		).
		Returning(querytest.TUsersID).
		Build()

	want := "INSERT INTO users (email, name) VALUES ($1, $2) ON CONFLICT (email) DO UPDATE SET name = $3 RETURNING users.id"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 3 {
		t.Fatalf("args = %v", args)
	}
}

func TestInsertFromSelectClearsSets(t *testing.T) {
	src := query.Select(querytest.TUsersID).From(querytest.TUsers)
	sql, args := query.Insert(querytest.TUsers).
		Set(query.Assign(querytest.TUsersName, "ignored")).
		FromSelect(src).
		Build()

	want := "INSERT INTO users SELECT users.id FROM users"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 0 {
		t.Fatalf("args = %v", args)
	}
}

func TestInsertSetClearsFromSelect(t *testing.T) {
	src := query.Select(querytest.TUsersID).From(querytest.TUsers)
	sql, args := query.Insert(querytest.TUsers).
		FromSelect(src).
		Set(query.Assign(querytest.TUsersName, "alice")).
		Build()

	want := "INSERT INTO users (name) VALUES ($1)"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 1 || args[0] != "alice" {
		t.Fatalf("args = %v", args)
	}
}

func TestInsertSetAny(t *testing.T) {
	sql, args := query.Insert(querytest.TUsers).
		SetAny("legacy_col", 99).
		Set(query.Assign(querytest.TUsersName, "alice")).
		Build()

	want := "INSERT INTO users (legacy_col, name) VALUES ($1, $2)"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if args[0] != 99 || args[1] != "alice" {
		t.Fatalf("args = %v", args)
	}
}

func TestInsertOnConflictDoNothing(t *testing.T) {
	sql, args := query.Insert(querytest.TUsers).
		Set(query.Assign(querytest.TUsersEmail, "a@b.com")).
		OnConflict(querytest.TUsersEmail).
		DoNothing().
		Build()

	want := "INSERT INTO users (email) VALUES ($1) ON CONFLICT (email) DO NOTHING"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 1 || args[0] != "a@b.com" {
		t.Fatalf("args = %v", args)
	}
}

func TestInsertPrefixSuffix(t *testing.T) {
	sql, args := query.Insert(querytest.TUsers).
		Prefix("/* batch */ ").
		Set(query.Assign(querytest.TUsersName, "alice")).
		Suffix(" ON CONFLICT DO NOTHING").
		Build()

	want := "/* batch */ INSERT INTO users (name) VALUES ($1) ON CONFLICT DO NOTHING"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 1 {
		t.Fatalf("args = %v", args)
	}
}

func TestInsertAssignExpr(t *testing.T) {
	sql, args := query.Insert(querytest.TUsers).
		Set(query.Assign(querytest.TUsersName, "alice"), query.AssignExpr(querytest.TUsersUpdatedAt, "NOW()")).
		Build()

	if want := "INSERT INTO users (name, updated_at) VALUES ($1, NOW())"; sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 1 || args[0] != "alice" {
		t.Fatalf("args = %v", args)
	}
}

func TestInsertSetNullBindsNil(t *testing.T) {
	sql, args := query.Insert(querytest.TUsers).
		Set(query.Assign(querytest.TUsersEmail, nil)).
		Build()

	if want := "INSERT INTO users (email) VALUES ($1)"; sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 1 || args[0] != nil {
		t.Fatalf("args = %v", args)
	}
}

func TestInsertOnConflictExcludedAndExpr(t *testing.T) {
	sql, args := query.Insert(querytest.TUsers).
		Set(query.Assign(querytest.TUsersID, int64(1)), query.Assign(querytest.TUsersName, "alice")).
		OnConflict(querytest.TUsersID).
		DoUpdate(
			query.AssignExcluded(querytest.TUsersName),
			query.AssignExpr(querytest.TUsersStatus, "COALESCE(?, users.status)", "x"),
		).
		Build()

	want := "INSERT INTO users (id, name) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, status = COALESCE($3, users.status)"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 3 || args[2] != "x" {
		t.Fatalf("args = %v", args)
	}
}
