package melody

import "testing"

func TestMelodyFactoryAppliesDialect(t *testing.T) {
	m := With(Postgres)

	q, _, _ := m.New("users").Where("id", "=", 1).Get()
	if q != "SELECT * FROM users WHERE id = $1" {
		t.Errorf("New: %q", q)
	}

	qi, _, _ := m.NewInsert("users").Set("name", "bob").Get()
	if qi != "INSERT INTO users (name) VALUES( $1 )" {
		t.Errorf("NewInsert: %q", qi)
	}

	qu, _, _ := m.NewUpdate("users").Set("name", "bob").Where("id", "=", 1).Get()
	if qu != "UPDATE users SET name = $1 WHERE id = $2" {
		t.Errorf("NewUpdate: %q", qu)
	}

	qd, _, _ := m.NewDelete("users").Where("id", "=", 1).Get()
	if qd != "DELETE FROM users WHERE id = $1" {
		t.Errorf("NewDelete: %q", qd)
	}
}

func TestMelodyFactoryPerQueryOverride(t *testing.T) {
	// a factory builder can still override the dialect per query
	q, _, _ := With(Postgres).New("users").Dialect(Default).Where("id", "=", 1).Get()
	if q != "SELECT * FROM users WHERE id = ?" {
		t.Errorf("override: %q", q)
	}
}
