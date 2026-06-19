package melody

import (
	"errors"
	"fmt"
	"strings"
)

type DeleteBuilder struct {
	ctx     deleteContext
	dialect Dialect
}

type deleteContext struct {
	Table     string
	Where     []WhereContext
	Returning []string
}

func NewDelete(table string) *DeleteBuilder {
	return &DeleteBuilder{
		ctx: deleteContext{
			Table: table,
		},
	}
}

// Dialect sets the SQL dialect used to render the final query.
func (d *DeleteBuilder) Dialect(dialect Dialect) *DeleteBuilder {
	d.dialect = dialect
	return d
}

// Returning adds a RETURNING clause (Postgres / SQLite syntax).
func (d *DeleteBuilder) Returning(columns ...string) *DeleteBuilder {
	d.ctx.Returning = append(d.ctx.Returning, columns...)
	return d
}

func (d *DeleteBuilder) Where(key string, operator string, values ...interface{}) *DeleteBuilder {
	return d.where(key, operator, values, false, false)
}

func (d *DeleteBuilder) OrWhere(key string, operator string, values ...interface{}) *DeleteBuilder {
	return d.where(key, operator, values, true, false)
}

func (d *DeleteBuilder) GroupWhere(sub SubBuilderFunc) *DeleteBuilder {
	return d.sub(sub, false)
}

func (d *DeleteBuilder) OrGroupWhere(sub SubBuilderFunc) *DeleteBuilder {
	return d.sub(sub, true)
}

func (d *DeleteBuilder) sub(sub SubBuilderFunc, isOr bool) *DeleteBuilder {
	wc := &WhereContext{IsOr: isOr}
	sub(wc)
	d.ctx.Where = append(d.ctx.Where, *wc)
	return d
}

func (d *DeleteBuilder) where(key, operator string, values []interface{}, isOr, isOn bool) *DeleteBuilder {
	d.ctx.Where = append(d.ctx.Where, WhereContext{
		Values: []where{
			{
				Key:      key,
				Operator: strings.ToUpper(operator),
				Values:   values,
				IsOr:     isOr,
				IsOn:     isOn,
			},
		},
	})
	return d
}

func (d *DeleteBuilder) Get() (query string, params []interface{}, err error) {
	query, params, err = d.build()
	if err != nil {
		return query, params, err
	}
	if d.dialect != nil {
		query = d.dialect.Rebind(query)
	}
	return query, params, err
}

func (d *DeleteBuilder) build() (res string, params []interface{}, err error) {
	if d.ctx.Table == "" {
		return res, params, errors.New("one table need to be defined")
	}

	result := []string{fmt.Sprintf("DELETE FROM %s", d.ctx.Table)}

	for i, wc := range d.ctx.Where {
		var r []string
		var p []interface{}

		r, p, err = buildWhere(wc, i == 0, false, false)
		if err != nil {
			return
		}

		result = append(result, r...)
		params = append(params, p...)
	}

	if len(d.ctx.Returning) != 0 {
		result = append(result, "RETURNING "+strings.Join(d.ctx.Returning, ", "))
	}

	return strings.Join(result, " "), params, nil
}
