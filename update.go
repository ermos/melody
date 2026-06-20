package melody

import (
	"errors"
	"fmt"
	"strings"
)

type UpdateBuilder struct {
	ctx     updateContext
	dialect Dialect
}

type updateContext struct {
	Table     string
	Value     []updateValue
	Where     []WhereContext
	Returning []string
}

type updateValue struct {
	Column string
	Value  any
	Raw    string // when set, RHS is this expression; Args are its bound params
	Args   []any
}

func NewUpdate(table string) *UpdateBuilder {
	return &UpdateBuilder{
		ctx: updateContext{
			Table: table,
		},
	}
}

// Dialect sets the SQL dialect used to render the final query.
func (u *UpdateBuilder) Dialect(d Dialect) *UpdateBuilder {
	u.dialect = d
	return u
}

func (u *UpdateBuilder) Set(column string, value any) *UpdateBuilder {
	u.ctx.Value = append(u.ctx.Value, updateValue{Column: column, Value: value})
	return u
}

// SetRaw sets a column to a raw expression instead of a bound value, e.g.
// SetRaw("views", "views + 1") or SetRaw("name", "COALESCE(?, name)", newName).
func (u *UpdateBuilder) SetRaw(column, expr string, args ...any) *UpdateBuilder {
	u.ctx.Value = append(u.ctx.Value, updateValue{Column: column, Raw: expr, Args: args})
	return u
}

// Returning adds a RETURNING clause (Postgres / SQLite syntax).
func (u *UpdateBuilder) Returning(columns ...string) *UpdateBuilder {
	u.ctx.Returning = append(u.ctx.Returning, columns...)
	return u
}

func (u *UpdateBuilder) Get() (query string, params []any, err error) {
	query, params, err = u.build()
	if u.dialect != nil {
		query = u.dialect.Rebind(query)
	}
	return query, params, err
}

func (u *UpdateBuilder) build() (res string, params []any, err error) {
	var result []string

	if u.ctx.Table == "" {
		return res, params, errors.New("one table need to be defined")
	}

	result = append(result, fmt.Sprintf("UPDATE %s SET", u.ctx.Table))

	var resultValue []string
	for _, v := range u.ctx.Value {
		if v.Raw != "" {
			resultValue = append(resultValue, fmt.Sprintf("%s = %s", v.Column, v.Raw))
			params = append(params, v.Args...)
		} else {
			resultValue = append(resultValue, fmt.Sprintf("%s = ?", v.Column))
			params = append(params, v.Value)
		}
	}

	result = append(result, strings.Join(resultValue, ", "))

	for i, wc := range u.ctx.Where {
		var r []string
		var p []any

		r, p, err = buildWhere(wc, i == 0, false, false)
		if err != nil {
			return
		}

		result = append(result, r...)
		params = append(params, p...)
	}

	if len(u.ctx.Returning) != 0 {
		result = append(result, "RETURNING "+strings.Join(u.ctx.Returning, ", "))
	}

	return strings.Join(result, " "), params, nil
}
