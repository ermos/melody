package melody

import (
	"errors"
	"fmt"
	"strings"
)

type InsertBuilder struct {
	ctx     insertContext
	dialect Dialect
}

type insertContext struct {
	Table          string
	Columns        []string
	Rows           [][]any
	DupKeys        []string // columns to overwrite on conflict
	ConflictTarget []string // unique columns the conflict is detected on (Postgres)
	DoNothing      bool     // emit ON CONFLICT DO NOTHING instead of an update
	Returning      []string
}

func NewInsert(table string) *InsertBuilder {
	return &InsertBuilder{
		ctx: insertContext{
			Table: table,
		},
	}
}

// Dialect sets the SQL dialect used to render the final query.
func (i *InsertBuilder) Dialect(d Dialect) *InsertBuilder {
	i.dialect = d
	return i
}

// Set adds a column/value to the current row. Column names are taken from the
// first row; subsequent rows (see AddRow) must Set the same columns in order.
func (i *InsertBuilder) Set(column string, value any) *InsertBuilder {
	if len(i.ctx.Rows) == 0 {
		i.ctx.Rows = append(i.ctx.Rows, nil)
	}
	last := len(i.ctx.Rows) - 1
	if last == 0 {
		i.ctx.Columns = append(i.ctx.Columns, column)
	}
	i.ctx.Rows[last] = append(i.ctx.Rows[last], value)
	return i
}

// AddRow starts a new row for a multi-row INSERT. The following Set calls fill
// it, in the same column order as the first row.
func (i *InsertBuilder) AddRow() *InsertBuilder {
	i.ctx.Rows = append(i.ctx.Rows, nil)
	return i
}

// Returning adds a RETURNING clause (Postgres / SQLite syntax).
func (i *InsertBuilder) Returning(columns ...string) *InsertBuilder {
	i.ctx.Returning = append(i.ctx.Returning, columns...)
	return i
}

// UpdateDuplicateKey flags the most recently Set column to be overwritten on
// conflict. The dialect decides the syntax: ON DUPLICATE KEY UPDATE (default)
// or ON CONFLICT ... DO UPDATE (Postgres — set the target with OnConflict).
// ponytail: uses the first row's value; combining with multi-row inserts is
// undefined — write VALUES(col)/EXCLUDED.col by hand if you need that.
func (i *InsertBuilder) UpdateDuplicateKey() *InsertBuilder {
	if n := len(i.ctx.Columns); n > 0 {
		i.ctx.DupKeys = append(i.ctx.DupKeys, i.ctx.Columns[n-1])
	}
	return i
}

// OnConflict sets the unique columns the conflict is detected on. Required by
// Postgres (ON CONFLICT (cols)); ignored by the default dialect.
func (i *InsertBuilder) OnConflict(columns ...string) *InsertBuilder {
	i.ctx.ConflictTarget = append(i.ctx.ConflictTarget, columns...)
	return i
}

// OnConflictDoNothing skips rows that violate a unique constraint
// (Postgres/SQLite: ON CONFLICT DO NOTHING; default/MySQL: a no-op
// ON DUPLICATE KEY UPDATE). Pair with OnConflict to target specific columns.
func (i *InsertBuilder) OnConflictDoNothing() *InsertBuilder {
	i.ctx.DoNothing = true
	return i
}

func (i *InsertBuilder) dia() Dialect {
	if i.dialect == nil {
		return Default
	}
	return i.dialect
}

func (i *InsertBuilder) Get() (query string, params []any, err error) {
	query, params, err = i.build()
	if err != nil {
		return query, params, err
	}
	if i.dialect != nil {
		query = i.dialect.Rebind(query)
	}
	return query, params, err
}

func (i *InsertBuilder) build() (res string, params []any, err error) {
	if i.ctx.Table == "" {
		return res, params, errors.New("one table need to be defined")
	}
	if len(i.ctx.Rows) == 0 {
		return res, params, errors.New("no values to insert")
	}

	n := len(i.ctx.Columns)
	result := []string{
		fmt.Sprintf("INSERT INTO %s", i.ctx.Table),
		fmt.Sprintf("(%s)", strings.Join(i.ctx.Columns, ", ")),
	}

	if len(i.ctx.Rows) == 1 {
		result = append(result, "VALUES(", placeholders(n, ", "), ")")
		params = append(params, i.ctx.Rows[0]...)
	} else {
		var rows []string
		for r, row := range i.ctx.Rows {
			if len(row) != n {
				return res, params, fmt.Errorf("row %d has %d values, expected %d", r, len(row), n)
			}
			rows = append(rows, "("+placeholders(n, ", ")+")")
			params = append(params, row...)
		}
		result = append(result, "VALUES "+strings.Join(rows, ", "))
	}

	if i.ctx.DoNothing {
		d, ok := i.dia().(doNothingDialect)
		if !ok {
			return res, params, errors.New("dialect does not support ON CONFLICT DO NOTHING")
		}
		result = append(result, d.ConflictDoNothing(i.ctx.ConflictTarget, i.ctx.Columns))
	} else if len(i.ctx.DupKeys) > 0 {
		clause, withParams := i.dia().UpsertClause(i.ctx.ConflictTarget, i.ctx.DupKeys)
		result = append(result, clause)
		if withParams {
			for _, col := range i.ctx.DupKeys {
				params = append(params, i.ctx.Rows[0][indexOf(i.ctx.Columns, col)])
			}
		}
	}

	if len(i.ctx.Returning) != 0 {
		result = append(result, "RETURNING "+strings.Join(i.ctx.Returning, ", "))
	}

	return strings.Join(result, " "), params, nil
}

func indexOf(ss []string, s string) int {
	for i, v := range ss {
		if v == s {
			return i
		}
	}
	return 0
}
