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
	Table     string
	Columns   []string
	Rows      [][]any
	DupKeys   []string // columns to update on ON DUPLICATE KEY UPDATE
	Returning []string
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

// UpdateDuplicateKey flags the most recently Set column for an
// ON DUPLICATE KEY UPDATE clause (MySQL).
// ponytail: dup-key uses the first row's value; combining it with multi-row
// inserts is undefined — use VALUES(col) by hand if you need that.
func (i *InsertBuilder) UpdateDuplicateKey() *InsertBuilder {
	if n := len(i.ctx.Columns); n > 0 {
		i.ctx.DupKeys = append(i.ctx.DupKeys, i.ctx.Columns[n-1])
	}
	return i
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

	if len(i.ctx.DupKeys) > 0 {
		var ups []string
		for _, col := range i.ctx.DupKeys {
			ups = append(ups, fmt.Sprintf("%s = ?", col))
			params = append(params, i.ctx.Rows[0][indexOf(i.ctx.Columns, col)])
		}
		result = append(result, "ON DUPLICATE KEY UPDATE")
		result = append(result, strings.Join(ups, ", "))
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
