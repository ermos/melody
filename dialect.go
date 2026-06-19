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
}

// Default keeps "?" placeholders (MySQL / SQLite style). It is the zero-value
// dialect so existing behaviour is unchanged when no dialect is set.
var Default Dialect = defaultDialect{}

// Postgres uses positional "$1, $2, ..." placeholders.
var Postgres Dialect = postgresDialect{}

type defaultDialect struct{}

func (defaultDialect) Rebind(query string) string { return query }

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
