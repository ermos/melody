package melody

import (
	"errors"
	"fmt"
	"strings"
)

type InsertBuilder struct {
	ctx insertContext
}

type insertContext struct {
	Table            string
	Value            []insertValue
	withDuplicateKey bool
}

type insertValue struct {
	Column           string
	Value            interface{}
	withDuplicateKey bool
}

func NewInsert(table string) *InsertBuilder {
	return &InsertBuilder{
		insertContext{
			Table: table,
		},
	}
}

func (i *InsertBuilder) Set(column string, value interface{}) *InsertBuilder {
	i.ctx.Value = append(i.ctx.Value, insertValue{Column: column, Value: value})
	return i
}

func (i *InsertBuilder) UpdateDuplicateKey() *InsertBuilder {
	i.ctx.withDuplicateKey = true
	i.ctx.Value[len(i.ctx.Value)-1].withDuplicateKey = true
	return i
}

func (i *InsertBuilder) Get() (query string, params []interface{}, err error) {
	return i.build()
}

func (i *InsertBuilder) build() (res string, params []interface{}, err error) {
	var result []string

	if i.ctx.Table == "" {
		return res, params, errors.New("one table need to be defined")
	}

	result = append(result, fmt.Sprintf("INSERT INTO %s", i.ctx.Table))

	var columns []string
	for _, v := range i.ctx.Value {
		columns = append(columns, v.Column)
	}

	result = append(result, fmt.Sprintf("(%s)", strings.Join(columns, ", ")))

	for _, v := range i.ctx.Value {
		params = append(params, v.Value)
	}

	result = append(result, "VALUES(")
	result = append(result, strings.Join(strings.Split(strings.Repeat("?", len(i.ctx.Value)), ""), ", "))
	result = append(result, ")")

	if i.ctx.withDuplicateKey {
		var resultOnUpdate []string
		for _, v := range i.ctx.Value {
			if v.withDuplicateKey {
				resultOnUpdate = append(resultOnUpdate, fmt.Sprintf("%s = ?", v.Column))
				params = append(params, v.Value)
			}
		}

		result = append(result, "ON DUPLICATE KEY UPDATE")
		result = append(result, strings.Join(resultOnUpdate, ", "))
	}

	return strings.Join(result, " "), params, nil
}
