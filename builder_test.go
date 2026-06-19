package melody

import (
	"reflect"
	"testing"
)

// Characterization tests: they pin the CURRENT behaviour of the builder
// (including known quirks) so later refactors are provably behaviour-preserving.
// Where a case documents a known bug, the comment says so.

func eq(t *testing.T, name, gotQuery string, gotParams []interface{}, gotErr error, wantQuery string, wantParams []interface{}) {
	t.Helper()
	if gotErr != nil {
		t.Fatalf("%s: unexpected error: %v", name, gotErr)
	}
	if gotQuery != wantQuery {
		t.Errorf("%s:\n got query:  %q\n want query: %q", name, gotQuery, wantQuery)
	}
	if len(gotParams) == 0 && len(wantParams) == 0 {
		return
	}
	if !reflect.DeepEqual(gotParams, wantParams) {
		t.Errorf("%s:\n got params:  %#v\n want params: %#v", name, gotParams, wantParams)
	}
}

func TestSelectStar(t *testing.T) {
	q, p, err := New("users").Get()
	eq(t, "star", q, p, err, "SELECT * FROM users", nil)
}

func TestSelectFields(t *testing.T) {
	q, p, err := New("users").Select("id", "name").Get()
	eq(t, "select fields", q, p, err, "SELECT id, name FROM users", nil)
}

func TestSelectDistinct(t *testing.T) {
	q, p, err := New("users").Distinct().Select("name").Get()
	eq(t, "distinct", q, p, err, "SELECT DISTINCT name FROM users", nil)
}

func TestWhereSingle(t *testing.T) {
	q, p, err := New("users").Where("id", "=", 1).Get()
	eq(t, "where single", q, p, err, "SELECT * FROM users WHERE id = ?", []interface{}{1})
}

func TestWhereAnd(t *testing.T) {
	q, p, err := New("users").Where("id", "=", 1).Where("name", "=", "bob").Get()
	eq(t, "where and", q, p, err, "SELECT * FROM users WHERE id = ? AND name = ?", []interface{}{1, "bob"})
}

func TestWhereOr(t *testing.T) {
	q, p, err := New("users").Where("id", "=", 1).OrWhere("id", "=", 2).Get()
	eq(t, "where or", q, p, err, "SELECT * FROM users WHERE id = ? OR id = ?", []interface{}{1, 2})
}

func TestWhereIn(t *testing.T) {
	q, p, err := New("users").Where("id", "IN", 1, 2, 3).Get()
	eq(t, "where in", q, p, err, "SELECT * FROM users WHERE id IN (?,?,?)", []interface{}{1, 2, 3})
}

func TestGroupWhere(t *testing.T) {
	q, p, err := New("users").
		Where("active", "=", true).
		GroupWhere(func(w *WhereContext) {
			w.Where("age", ">", 18).OrWhere("vip", "=", true)
		}).Get()
	eq(t, "group where", q, p, err,
		"SELECT * FROM users WHERE active = ? AND ( age > ? OR vip = ? )",
		[]interface{}{true, 18, true})
}

func TestOrderByGroupByLimitOffset(t *testing.T) {
	q, p, err := New("users").
		Where("active", "=", true).
		GroupBy("country").
		OrderBy("name", Asc).
		Limit(10).
		Offset(20).
		Get()
	eq(t, "full", q, p, err,
		"SELECT * FROM users WHERE active = ? GROUP BY country ORDER BY name ASC LIMIT 10 OFFSET 20",
		[]interface{}{true})
}

func TestJoin(t *testing.T) {
	q, p, err := New("users").
		LeftJoin("orders", func(w *WhereContext) {
			w.On("orders.user_id", "=", "users.id")
		}).Get()
	eq(t, "join", q, p, err,
		"SELECT * FROM users LEFT JOIN orders ON orders.user_id = users.id", nil)
}

func TestCount(t *testing.T) {
	q, p, err := New("users").Where("active", "=", true).GetCount()
	eq(t, "count", q, p, err, "SELECT count(*) FROM users WHERE active = ?", []interface{}{true})
}

func TestCountWithKey(t *testing.T) {
	q, p, err := New("users").GetCountWithKey("id")
	eq(t, "count key", q, p, err, "SELECT count(id) FROM users", nil)
}

func TestNoTableErrors(t *testing.T) {
	_, _, err := New().Get()
	if err == nil {
		t.Fatal("expected error when no table defined")
	}
}

func TestInsert(t *testing.T) {
	q, p, err := NewInsert("users").Set("name", "bob").Set("age", 30).Get()
	eq(t, "insert", q, p, err,
		"INSERT INTO users (name, age) VALUES( ?, ? )", []interface{}{"bob", 30})
}

func TestUpdate(t *testing.T) {
	q, p, err := NewUpdate("users").Set("name", "bob").Where("id", "=", 1).Get()
	eq(t, "update", q, p, err,
		"UPDATE users SET name = ? WHERE id = ?", []interface{}{"bob", 1})
}
