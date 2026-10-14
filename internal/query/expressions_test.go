package query_test

import (
	"testing"

	"github.com/yaziciahmet/typg/internal/query"
	"github.com/yaziciahmet/typg/internal/query/querytest"
)

type aliasTable struct{ name, alias string }

func (a aliasTable) SQLName() string { return a.name }
func (a aliasTable) Alias() string   { return a.alias }

func TestColumnAsInEachClause(t *testing.T) {
	uid := query.ColumnAs(querytest.TUsersID, "uid")

	sql, args := query.Select(uid, querytest.TUsersName).
		From(querytest.TUsers).
		Where(query.Eq(uid, int64(1))).
		GroupBy(uid).
		OrderByDesc(uid).
		Build()

	want := "SELECT users.id AS uid, users.name FROM users WHERE users.id = $1 GROUP BY users.id ORDER BY uid DESC"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 1 {
		t.Fatalf("args = %v", args)
	}
}

func TestColumnAsInReturning(t *testing.T) {
	sql, _ := query.Insert(querytest.TUsers).
		Set(query.Assign(querytest.TUsersName, "a")).
		Returning(query.ColumnAs(querytest.TUsersID, "new_id")).
		Build()

	if want := "INSERT INTO users (name) VALUES ($1) RETURNING users.id AS new_id"; sql != want {
		t.Fatalf("sql = %q", sql)
	}
}

func TestAggregateHelpersWithAliasHavingAndOrder(t *testing.T) {
	cnt := query.Count().As("cnt")

	sql, args := query.Select(querytest.TUsersStatus, cnt, query.Max(querytest.TUsersID).As("last_id")).
		From(querytest.TUsers).
		GroupBy(querytest.TUsersStatus).
		Having(query.Gt(query.Count(), 1)).
		OrderByDesc(cnt).
		Build()

	want := "SELECT users.status, count(*) AS cnt, max(users.id) AS last_id FROM users GROUP BY users.status HAVING count(*) > $1 ORDER BY cnt DESC"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 1 || args[0] != 1 {
		t.Fatalf("args = %v", args)
	}
}

func TestAggregateHelpers(t *testing.T) {
	sql, args := query.Select(
		query.CountOf(querytest.TUsersEmail),
		query.Sum(querytest.TUsersID),
		query.Min(querytest.TUsersID),
		query.Coalesce(querytest.TUsersName, "anon").As("display"),
		query.Coalesce(querytest.TUsersName, querytest.TUsersEmail),
	).From(querytest.TUsers).Build()

	want := "SELECT count(users.email), sum(users.id), min(users.id), COALESCE(users.name, $1) AS display, COALESCE(users.name, users.email) FROM users"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 1 || args[0] != "anon" {
		t.Fatalf("args = %v", args)
	}
}

func TestPredicateWithRawColumnBindsItsArgs(t *testing.T) {
	sql, args := query.Select(querytest.TUsersID).From(querytest.TUsers).
		Where(query.Eq(query.Coalesce(querytest.TUsersName, "anon"), "bob")).
		Build()

	if want := "SELECT users.id FROM users WHERE COALESCE(users.name, $1) = $2"; sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 2 || args[0] != "anon" || args[1] != "bob" {
		t.Fatalf("args = %v", args)
	}
}

func TestNotNotInBetweenEqAny(t *testing.T) {
	sql, args := query.Select(querytest.TUsersID).From(querytest.TUsers).
		Where(query.Not(query.Eq(querytest.TUsersName, "a"))).
		Where(query.NotIn(querytest.TUsersID, int64(1), int64(2))).
		Where(query.Between(querytest.TUsersID, int64(10), int64(20))).
		Where(query.EqAny(querytest.TUsersEmail, []string{"x", "y"})).
		Build()

	want := "SELECT users.id FROM users WHERE NOT (users.name = $1) AND users.id NOT IN ($2, $3) AND users.id BETWEEN $4 AND $5 AND users.email = ANY($6)"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 6 {
		t.Fatalf("args = %v", args)
	}
}

func TestNotInEmptyRendersTrue(t *testing.T) {
	sql, _ := query.Select(querytest.TUsersID).From(querytest.TUsers).Where(query.NotIn(querytest.TUsersID)).Build()
	if want := "SELECT users.id FROM users WHERE TRUE"; sql != want {
		t.Fatalf("sql = %q", sql)
	}
}

func TestExistsAndInSelectKeepPlaceholderNumbering(t *testing.T) {
	sub := query.Select(querytest.TUsersID).From(querytest.TUsers).Where(query.Eq(querytest.TUsersName, "inner"))

	sql, args := query.Select(querytest.TUsersID).From(querytest.TUsers).
		Where(query.Eq(querytest.TUsersEmail, "outer")).
		Where(query.InSelect(querytest.TUsersID, sub)).
		Where(query.Not(query.Exists(sub))).
		Build()

	want := "SELECT users.id FROM users WHERE users.email = $1 AND users.id IN (SELECT users.id FROM users WHERE users.name = $2) AND NOT (EXISTS (SELECT users.id FROM users WHERE users.name = $3))"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 3 || args[0] != "outer" || args[1] != "inner" || args[2] != "inner" {
		t.Fatalf("args = %v", args)
	}
}

func TestAliasedTableInStatements(t *testing.T) {
	u := aliasTable{name: "users", alias: "u"}

	sql, _ := query.Update(u).Set(query.Assign(querytest.TUsersName, "x")).Where(query.Eq(query.Col("u", "id"), int64(1))).Build()
	if want := "UPDATE users AS u SET name = $1 WHERE u.id = $2"; sql != want {
		t.Fatalf("update = %q", sql)
	}

	sql, _ = query.Delete(u).Where(query.Eq(query.Col("u", "id"), int64(1))).Build()
	if want := "DELETE FROM users AS u WHERE u.id = $1"; sql != want {
		t.Fatalf("delete = %q", sql)
	}

	sql, _ = query.Insert(u).Set(query.Assign(querytest.TUsersName, "x")).Build()
	if want := "INSERT INTO users AS u (name) VALUES ($1)"; sql != want {
		t.Fatalf("insert = %q", sql)
	}

	e := aliasTable{name: "users", alias: "e"}
	sql, _ = query.Select(query.Col("e", "name"), query.ColumnAs(query.Col("u", "name"), "manager")).
		From(e).
		Join(u, query.EqCol(query.Col("e", "id"), query.Col("u", "id"))).
		Build()
	if want := "SELECT e.name, u.name AS manager FROM users AS e INNER JOIN users AS u ON e.id = u.id"; sql != want {
		t.Fatalf("self join = %q", sql)
	}
}
