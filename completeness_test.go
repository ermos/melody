package melody

import (
	"testing"
	"time"
)

func TestDelete(t *testing.T) {
	q, p, err := NewDelete("users").Where("id", "=", 1).Get()
	eq(t, "delete", q, p, err, "DELETE FROM users WHERE id = ?", []interface{}{1})
}

func TestDeleteNoTable(t *testing.T) {
	_, _, err := NewDelete("").Get()
	if err == nil {
		t.Fatal("expected error when no table defined")
	}
}

func TestDeletePostgresReturning(t *testing.T) {
	q, p, err := NewDelete("users").Dialect(Postgres).
		Where("id", "=", 1).Returning("id").Get()
	eq(t, "delete returning", q, p, err,
		"DELETE FROM users WHERE id = $1 RETURNING id", []interface{}{1})
}

func TestInsertReturning(t *testing.T) {
	q, p, err := NewInsert("users").Dialect(Postgres).
		Set("name", "bob").Returning("id", "created_at").Get()
	eq(t, "insert returning", q, p, err,
		"INSERT INTO users (name) VALUES( $1 ) RETURNING id, created_at",
		[]interface{}{"bob"})
}

func TestUpdateReturning(t *testing.T) {
	q, p, err := NewUpdate("users").
		Set("name", "bob").Where("id", "=", 1).Returning("updated_at").Get()
	eq(t, "update returning", q, p, err,
		"UPDATE users SET name = ? WHERE id = ? RETURNING updated_at",
		[]interface{}{"bob", 1})
}

func TestParseStructWithTimeAndNested(t *testing.T) {
	type Address struct {
		City string `json:"city" db:"city"`
	}
	type User struct {
		ID        int       `json:"id" db:"id"`
		CreatedAt time.Time `json:"created_at" db:"created_at"`
		Address   Address   `json:"address"`
		secret    string    // unexported, must be ignored
	}

	b := New("users")
	b.parseStruct(User{})

	want := map[string]string{
		"id":           "id",
		"created_at":   "created_at",   // time.Time mapped, not recursed
		"address.city": "city",          // nested struct prefixed
	}
	for k, v := range want {
		if got := b.ctx.JsonToDB[k]; got != v {
			t.Errorf("JsonToDB[%q] = %q, want %q", k, got, v)
		}
	}
	if len(b.ctx.JsonToDB) != len(want) {
		t.Errorf("unexpected extra mappings: %#v", b.ctx.JsonToDB)
	}
}
