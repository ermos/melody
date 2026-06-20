package melody

import (
	"strconv"
	"strings"
)

// Dialect adapts the generated SQL to a specific database.
//
// The builder always emits parameters as "?" placeholders internally; the
// dialect's Rebind translates them to the engine's native style as the final
// step. Since the builder never emits string literals (every value is a bound
// parameter), a "?" can only ever be a placeholder, which makes the rebind
// pass safe.
type Dialect interface {
	// Rebind converts the builder's "?" placeholders into the dialect's
	// native placeholder style.
	Rebind(query string) string
	// UpsertClause renders the "update on conflict" clause of an INSERT.
	// conflict holds the columns forming the unique constraint (Postgres needs
	// them; MySQL ignores them); update holds the columns to overwrite.
	// withParams reports whether the caller must append one bound value per
	// update column (MySQL "col = ?"); Postgres uses EXCLUDED and needs none.
	UpsertClause(conflict, update []string) (clause string, withParams bool)
}

// Default keeps "?" placeholders (MySQL / SQLite style). It is the zero-value
// dialect so existing behaviour is unchanged when no dialect is set.
var Default Dialect = defaultDialect{}

// Postgres uses positional "$1, $2, ..." placeholders.
var Postgres Dialect = postgresDialect{}

// SQLite keeps "?" placeholders but uses the Postgres-style
// ON CONFLICT ... DO UPDATE SET upsert (not MySQL's ON DUPLICATE KEY).
var SQLite Dialect = sqliteDialect{}

// doNothingDialect is an optional capability: a dialect that can render an
// "insert, skip on conflict" clause. Detected via type assertion so the public
// Dialect interface stays unchanged (custom dialects need not implement it).
type doNothingDialect interface {
	ConflictDoNothing(conflict, columns []string) string
}

type defaultDialect struct{}

func (defaultDialect) Rebind(query string) string { return query }

func (defaultDialect) UpsertClause(_, update []string) (string, bool) {
	sets := make([]string, len(update))
	for i, c := range update {
		sets[i] = c + " = ?"
	}
	return "ON DUPLICATE KEY UPDATE " + strings.Join(sets, ", "), true
}

// ConflictDoNothing: MySQL has no DO NOTHING, so emit a no-op self-assignment.
func (defaultDialect) ConflictDoNothing(_, columns []string) string {
	if len(columns) == 0 {
		return "ON DUPLICATE KEY UPDATE"
	}
	c := columns[0]
	return "ON DUPLICATE KEY UPDATE " + c + " = " + c
}

type postgresDialect struct{}

func (postgresDialect) Rebind(query string) string {
	var sb strings.Builder
	sb.Grow(len(query))
	n := 0
	for i := 0; i < len(query); i++ {
		if query[i] == '?' {
			n++
			sb.WriteByte('$')
			sb.WriteString(strconv.Itoa(n))
		} else {
			sb.WriteByte(query[i])
		}
	}
	return sb.String()
}

func (postgresDialect) UpsertClause(conflict, update []string) (string, bool) {
	return onConflictClause(conflict, update)
}

func (postgresDialect) ConflictDoNothing(conflict, _ []string) string {
	return doNothingClause(conflict)
}

type sqliteDialect struct{}

func (sqliteDialect) Rebind(query string) string { return query }

func (sqliteDialect) UpsertClause(conflict, update []string) (string, bool) {
	return onConflictClause(conflict, update)
}

func (sqliteDialect) ConflictDoNothing(conflict, _ []string) string {
	return doNothingClause(conflict)
}

func doNothingClause(conflict []string) string {
	target := ""
	if len(conflict) > 0 {
		target = " (" + strings.Join(conflict, ", ") + ")"
	}
	return "ON CONFLICT" + target + " DO NOTHING"
}

// onConflictClause renders the Postgres/SQLite upsert clause. It needs no extra
// params (values come from EXCLUDED).
func onConflictClause(conflict, update []string) (string, bool) {
	sets := make([]string, len(update))
	for i, c := range update {
		sets[i] = c + " = EXCLUDED." + c
	}
	target := ""
	if len(conflict) > 0 {
		target = " (" + strings.Join(conflict, ", ") + ")"
	}
	return "ON CONFLICT" + target + " DO UPDATE SET " + strings.Join(sets, ", "), false
}
