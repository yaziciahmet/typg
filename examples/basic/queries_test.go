package basic_test

import (
	"testing"

	"github.com/yaziciahmet/typg/examples/basic/gen/db"
	"github.com/yaziciahmet/typg/examples/basic/gen/models"
)

func TestSelectJoin(t *testing.T) {
	sql, args := db.Select(db.Users.Name, db.Posts.Title).
		From(db.Users).
		Join(db.Posts, db.Users.ID.EqCol(db.Posts.UserID)).
		Where(db.Users.Status.Eq(models.UserStatusActive)).
		Where(db.Posts.Published.Eq(true)).
		Build()

	want := "SELECT users.name, posts.title FROM users INNER JOIN posts ON users.id = posts.user_id WHERE users.status = $1 AND posts.published = $2"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 2 {
		t.Fatalf("args = %v", args)
	}
	if args[0] != models.UserStatusActive || args[1] != true {
		t.Fatalf("args = %v", args)
	}
}

func TestArrayContains(t *testing.T) {
	sql, args := db.Select(db.Users.ID).
		From(db.Users).
		Where(db.Users.Tags.Contains("go")).
		Build()

	want := "SELECT users.id FROM users WHERE users.tags @> ARRAY[$1]"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 1 || args[0] != "go" {
		t.Fatalf("args = %v", args)
	}
}

func TestStringILike(t *testing.T) {
	sql, args := db.Select(db.Users.ID).
		From(db.Users).
		Where(db.Users.Email.ILike("%@example.com")).
		Build()

	want := "SELECT users.id FROM users WHERE users.email ILIKE $1"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if args[0] != "%@example.com" {
		t.Fatalf("args = %v", args)
	}
}

func TestSelectRaw(t *testing.T) {
	sql, args := db.Select(
		db.Users.ID,
		db.NewRawCol("COALESCE(?, ?) AS display", db.Users.Name, db.Users.Email),
	).
		From(db.Users).
		Build()

	want := "SELECT users.id, COALESCE(users.name, users.email) AS display FROM users"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 0 {
		t.Fatalf("args = %v", args)
	}
}

func TestTypedAssignments(t *testing.T) {
	sql, args := db.Insert(db.Users).
		Set(
			db.Users.Email.Set("a@b.com"),
			db.Users.Name.Set("alice"),
			db.Users.Status.Set(models.UserStatusActive),
		).
		OnConflict(db.Users.Email).
		DoUpdate(db.Users.Name.SetExcluded(), db.Users.UpdatedAt.SetExpr("now()")).
		Build()

	want := "INSERT INTO users (email, name, status) VALUES ($1, $2, $3) ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name, updated_at = now()"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 3 {
		t.Fatalf("args = %v", args)
	}

	sql, _ = db.Update(db.Users).
		Set(db.Users.LastLoginAt.SetNull(), db.Users.Name.Set("bob")).
		Where(db.Users.ID.Eq(1)).
		Build()

	want = "UPDATE users SET last_login_at = $1, name = $2 WHERE users.id = $3"
	if sql != want {
		t.Fatalf("sql = %q", sql)
	}
}
