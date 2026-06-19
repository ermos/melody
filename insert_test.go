package melody

import "testing"

func TestInsertMultiRow(t *testing.T) {
	q, p, err := NewInsert("users").
		Set("name", "bob").Set("age", 30).
		AddRow().
		Set("name", "alice").Set("age", 25).
		Get()
	eq(t, "multi row", q, p, err,
		"INSERT INTO users (name, age) VALUES (?, ?), (?, ?)",
		[]any{"bob", 30, "alice", 25})
}

func TestInsertMultiRowPostgres(t *testing.T) {
	q, p, err := NewInsert("users").Dialect(Postgres).
		Set("name", "bob").
		AddRow().Set("name", "alice").
		Get()
	eq(t, "multi row pg", q, p, err,
		"INSERT INTO users (name) VALUES ($1), ($2)",
		[]any{"bob", "alice"})
}

func TestInsertMultiRowMismatch(t *testing.T) {
	_, _, err := NewInsert("users").
		Set("name", "bob").Set("age", 30).
		AddRow().Set("name", "alice"). // missing age
		Get()
	if err == nil {
		t.Fatal("expected error on row value count mismatch")
	}
}

func TestInsertNoValues(t *testing.T) {
	_, _, err := NewInsert("users").Get()
	if err == nil {
		t.Fatal("expected error when no values set")
	}
}

func TestInsertDuplicateKey(t *testing.T) {
	q, p, err := NewInsert("users").
		Set("id", 1).
		Set("name", "bob").UpdateDuplicateKey().
		Get()
	eq(t, "dup key", q, p, err,
		"INSERT INTO users (id, name) VALUES( ?, ? ) ON DUPLICATE KEY UPDATE name = ?",
		[]any{1, "bob", "bob"})
}
