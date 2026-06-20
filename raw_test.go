package melody

import "testing"

func TestWhereNull(t *testing.T) {
	q, p, err := New("users").Where("active", "=", true).WhereNull("deleted_at").Get()
	eq(t, "is null", q, p, err,
		"SELECT * FROM users WHERE active = ? AND deleted_at IS NULL", []any{true})
}

func TestWhereNotNullOr(t *testing.T) {
	q, p, err := New("users").Where("id", "=", 1).OrWhereNotNull("email").Get()
	eq(t, "is not null or", q, p, err,
		"SELECT * FROM users WHERE id = ? OR email IS NOT NULL", []any{1})
}

func TestWhereRaw(t *testing.T) {
	q, p, err := New("users").WhereRaw("LOWER(name) LIKE LOWER(?)", "%bob%").Get()
	eq(t, "raw like", q, p, err,
		"SELECT * FROM users WHERE LOWER(name) LIKE LOWER(?)", []any{"%bob%"})
}

func TestWhereRawExistsAndBind(t *testing.T) {
	q, p, err := New("users").
		Where("active", "=", true).
		WhereRaw("EXISTS (SELECT 1 FROM orders WHERE orders.user_id = users.id AND total > ?)", 100).
		Get()
	eq(t, "raw exists", q, p, err,
		"SELECT * FROM users WHERE active = ? AND EXISTS (SELECT 1 FROM orders WHERE orders.user_id = users.id AND total > ?)",
		[]any{true, 100})
}

func TestRawInsideGroup(t *testing.T) {
	q, p, err := New("users").
		Where("active", "=", true).
		GroupWhere(func(w *WhereContext) {
			w.WhereNull("a").OrWhereRaw("b > ?", 5)
		}).Get()
	eq(t, "raw in group", q, p, err,
		"SELECT * FROM users WHERE active = ? AND ( a IS NULL OR b > ? )", []any{true, 5})
}

func TestOrderByRaw(t *testing.T) {
	q, _, err := New("users").OrderByRaw("RANDOM()").Get()
	if err != nil || q != "SELECT * FROM users ORDER BY RANDOM()" {
		t.Errorf("order by raw: %q (err=%v)", q, err)
	}
}

func TestUpdateSetRaw(t *testing.T) {
	q, p, err := NewUpdate("posts").
		SetRaw("views", "views + 1").
		Set("title", "hi").
		SetRaw("name", "COALESCE(?, name)", "new").
		Where("id", "=", 1).Get()
	eq(t, "set raw", q, p, err,
		"UPDATE posts SET views = views + 1, title = ?, name = COALESCE(?, name) WHERE id = ?",
		[]any{"hi", "new", 1})
}

func TestPostgresDoNothing(t *testing.T) {
	q, p, err := NewInsert("users").Dialect(Postgres).
		Set("id", 1).Set("name", "bob").
		OnConflict("id").OnConflictDoNothing().Get()
	eq(t, "pg do nothing", q, p, err,
		"INSERT INTO users (id, name) VALUES( $1, $2 ) ON CONFLICT (id) DO NOTHING",
		[]any{1, "bob"})
}

func TestSQLiteDoNothingNoTarget(t *testing.T) {
	q, p, err := NewInsert("users").Dialect(SQLite).
		Set("name", "bob").OnConflictDoNothing().Get()
	eq(t, "sqlite do nothing", q, p, err,
		"INSERT INTO users (name) VALUES( ? ) ON CONFLICT DO NOTHING", []any{"bob"})
}

func TestDefaultDoNothingMySQL(t *testing.T) {
	q, p, err := NewInsert("users").
		Set("id", 1).Set("name", "bob").OnConflictDoNothing().Get()
	eq(t, "mysql do nothing", q, p, err,
		"INSERT INTO users (id, name) VALUES( ?, ? ) ON DUPLICATE KEY UPDATE id = id", []any{1, "bob"})
}
