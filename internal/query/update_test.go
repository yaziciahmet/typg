package query_test

import (
	"testing"

	"github.com/yaziciahmet/typg/internal/query"
	"github.com/yaziciahmet/typg/internal/query/querytest"
)

func TestUpdateSetExpr(t *testing.T) {
	sql, args := query.Update(querytest.TUsers).
		Set(query.Assign(querytest.TUsersStatus, "active")).
		Set(query.AssignExpr(querytest.TUsersUpdatedAt, "NOW()")).
		Set(query.AssignExpr(querytest.TUsersName, "COALESCE(?, name)", "fallback")).
		Where(query.Eq(querytest.TUsersID, int64(1))).
		Build()

	want := "UPDATE users SET status = $1, updated_at = NOW(), name = COALESCE($2, name) WHERE users.id = $3"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 3 || args[0] != "active" || args[1] != "fallback" || args[2] != int64(1) {
		t.Fatalf("args = %v", args)
	}
}

func TestUpdateReturningFromSetOverwrite(t *testing.T) {
	sql, args := query.Update(querytest.TUsers).
		Set(query.Assign(querytest.TUsersStatus, "draft")).
		Set(query.Assign(querytest.TUsersStatus, "active")).
		Set(query.Assign(querytest.TUsersName, "bob")).
		From(query.Named("accounts")).
		Returning(querytest.TUsersID, querytest.TUsersName).
		Where(query.Ne(querytest.TUsersStatus, "deleted")).
		Build()

	want := "UPDATE users SET status = $1, name = $2 FROM accounts WHERE users.status != $3 RETURNING users.id, users.name"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 3 || args[0] != "active" || args[1] != "bob" || args[2] != "deleted" {
		t.Fatalf("args = %v", args)
	}
}

func TestUpdateSetAny(t *testing.T) {
	sql, args := query.Update(querytest.TUsers).
		SetAny("metadata", `{"k":1}`).
		Where(query.Eq(querytest.TUsersID, int64(5))).
		Build()

	want := "UPDATE users SET metadata = $1 WHERE users.id = $2"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if args[0] != `{"k":1}` || args[1] != int64(5) {
		t.Fatalf("args = %v", args)
	}
}

func TestUpdateFromTables(t *testing.T) {
	sql, args := query.Update(querytest.TUsers).
		Set(query.Assign(querytest.TUsersStatus, "x")).
		From(query.Named("a"), query.Named("b")).
		Where(query.EqCol(querytest.TUsersID, query.Col("a", "id"))).
		Build()

	if want := "UPDATE users SET status = $1 FROM a, b WHERE users.id = a.id"; sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 1 {
		t.Fatalf("args = %v", args)
	}
}

func TestUpdateSetManyAssignmentsAtOnce(t *testing.T) {
	sql, _ := query.Update(querytest.TUsers).
		Set(
			query.Assign(querytest.TUsersStatus, "x"),
			query.AssignExpr(querytest.TUsersName, "upper(?)", querytest.TUsersEmail),
			query.Assign(querytest.TUsersStatus, "y"),
		).
		Build()

	if want := "UPDATE users SET status = $1, name = upper(users.email)"; sql != want {
		t.Fatalf("sql = %q", sql)
	}
}
