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
	Table string
	Value []updateValue
	Where []WhereContext
}

type updateValue struct {
	Column string
	Value  interface{}
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

func (u *UpdateBuilder) Set(column string, value interface{}) *UpdateBuilder {
	u.ctx.Value = append(u.ctx.Value, updateValue{Column: column, Value: value})
	return u
}

func (u *UpdateBuilder) Get() (query string, params []interface{}, err error) {
	query, params, err = u.build()
	if u.dialect != nil {
		query = u.dialect.Rebind(query)
	}
	return query, params, err
}

func (u *UpdateBuilder) build() (res string, params []interface{}, err error) {
	var result []string

	if u.ctx.Table == "" {
		return res, params, errors.New("one table need to be defined")
	}

	result = append(result, fmt.Sprintf("UPDATE %s SET", u.ctx.Table))

	var resultValue []string
	for _, v := range u.ctx.Value {
		resultValue = append(resultValue, fmt.Sprintf("%s = ?", v.Column))
		params = append(params, v.Value)
	}

	result = append(result, strings.Join(resultValue, ", "))

	for i, wc := range u.ctx.Where {
		var r []string
		var p []interface{}

		r, p, err = buildWhere(wc, i == 0, false, false)
		if err != nil {
			return
		}

		result = append(result, r...)
		params = append(params, p...)
	}

	return strings.Join(result, " "), params, nil
}
