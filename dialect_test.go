package melody

import "testing"

func TestPostgresSelectPlaceholders(t *testing.T) {
	q, p, err := New("users").
		Dialect(Postgres).
		Where("id", "=", 1).
		Where("name", "=", "bob").
		Where("status", "IN", "a", "b").
		Get()
	eq(t, "pg select", q, p, err,
		"SELECT * FROM users WHERE id = $1 AND name = $2 AND status IN ($3,$4)",
		[]any{1, "bob", "a", "b"})
}

func TestPostgresInsert(t *testing.T) {
	q, p, err := NewInsert("users").Dialect(Postgres).
		Set("name", "bob").Set("age", 30).Get()
	eq(t, "pg insert", q, p, err,
		"INSERT INTO users (name, age) VALUES( $1, $2 )", []any{"bob", 30})
}

func TestPostgresUpdate(t *testing.T) {
	q, p, err := NewUpdate("users").Dialect(Postgres).
		Set("name", "bob").Where("id", "=", 1).Get()
	eq(t, "pg update", q, p, err,
		"UPDATE users SET name = $1 WHERE id = $2", []any{"bob", 1})
}

func TestDefaultDialectUnchanged(t *testing.T) {
	q, _, _ := New("users").Where("id", "=", 1).Get()
	if q != "SELECT * FROM users WHERE id = ?" {
		t.Errorf("default dialect changed: %q", q)
	}
}

func TestRebindDollar(t *testing.T) {
	got := Postgres.Rebind("a = ? AND b IN (?,?)")
	want := "a = $1 AND b IN ($2,$3)"
	if got != want {
		t.Errorf("Rebind = %q, want %q", got, want)
	}
}
