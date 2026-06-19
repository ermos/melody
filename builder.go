package melody

import (
	"errors"
	"fmt"
	"strings"
)

type SubBuilderFunc func(w *WhereContext)

type Builder struct {
	ctx     builderContext
	dialect Dialect
}

type builderContext struct {
	Table        []string
	Select       []string
	Join         []join
	Where        []WhereContext
	GroupBy      []string
	OrderBy      []orderBy
	Limit        int
	Offset       int
	WithDistinct bool
	JsonToDB     map[string]string
}

// WhereContext allows to embed multiple where/on for simulate OR condition
type WhereContext struct {
	Sub    []WhereContext
	Values []where
	IsOr   bool
}

type where struct {
	Key      string
	Operator string
	Values   []any
	IsOn     bool
	IsOr     bool
}

type orderBy struct {
	Key       string
	Direction string
}

func New(tables ...string) *Builder {
	b := Builder{}
	b.ctx.Table = append(b.ctx.Table, tables...)
	return &b
}

// Dialect sets the SQL dialect used to render the final query. Defaults to
// melody.Default ("?" placeholders) when unset.
func (b *Builder) Dialect(d Dialect) *Builder {
	b.dialect = d
	return b
}

func (b *Builder) rebind(query string) string {
	if b.dialect == nil {
		return query
	}
	return b.dialect.Rebind(query)
}

func (b *Builder) Select(fields ...string) *Builder {
	b.ctx.Select = append(b.ctx.Select, fields...)
	return b
}

func (b *Builder) Table(tables ...string) *Builder {
	b.ctx.Table = append(b.ctx.Table, tables...)
	return b
}

func (b *Builder) Distinct() *Builder {
	b.ctx.WithDistinct = true
	return b
}

func (b *Builder) GroupBy(fields ...string) *Builder {
	b.ctx.GroupBy = append(b.ctx.GroupBy, fields...)
	return b
}

func (b *Builder) OrderBy(field, direction string) *Builder {
	b.ctx.OrderBy = append(b.ctx.OrderBy, orderBy{
		Key:       field,
		Direction: direction,
	})
	return b
}

func (b *Builder) Limit(limit int) *Builder {
	b.ctx.Limit = limit
	return b
}

func (b *Builder) Offset(offset int) *Builder {
	b.ctx.Offset = offset
	return b
}

func (b *Builder) Get() (query string, params []any, err error) {
	query, params, err = b.build(true)
	return b.rebind(b.withFields(b.withLimitOffset(query))), params, err
}

func (b *Builder) GetCount() (query string, params []any, err error) {
	query, params, err = b.build(false)
	return b.rebind(b.withCount(query, "")), params, err
}

func (b *Builder) GetCountWithKey(key string) (query string, params []any, err error) {
	query, params, err = b.build(false)
	return b.rebind(b.withCount(query, key)), params, err
}

func (b *Builder) GetOffset() int {
	return b.ctx.Offset
}

func (b *Builder) GetLimit() int {
	return b.ctx.Limit
}

func buildWhere(wc WhereContext, isFirst, isJoin, isSub bool) (result []string, params []any, err error) {
	valuesLen := len(wc.Values)

	if isFirst && !isJoin && valuesLen != 0 {
		result = append(result, "WHERE")
	}

	if isSub || valuesLen > 1 {
		if !isFirst {
			if wc.IsOr {
				result = append(result, "OR")
			} else {
				result = append(result, "AND")
			}
		}
		result = append(result, "(")
	}

	for i, w := range wc.Values {
		if i != 0 || valuesLen == 1 && !isFirst {
			if w.IsOr {
				result = append(result, "OR")
			} else {
				result = append(result, "AND")
			}
		}

		if len(w.Values) != 1 && w.Operator != "IN" {
			return result, params, fmt.Errorf("%s cannot contains multiple value if operator isn't IN", w.Key)
		}

		if w.IsOn {
			result = append(result, fmt.Sprintf("%s %s %s", w.Key, w.Operator, w.Values[0].(string)))
		} else {
			if w.Operator == "IN" {
				result = append(result, fmt.Sprintf(
					"%s IN (%s)",
					w.Key,
					placeholders(len(w.Values), ","),
				))
			} else {
				result = append(result, fmt.Sprintf("%s %s ?", w.Key, w.Operator))
			}

			params = append(params, w.Values...)
		}
	}

	for i, s := range wc.Sub {
		var r []string
		var p []any

		r, p, err = buildWhere(s, isFirst && i == 0 && len(wc.Values) == 0, isJoin, true)
		if err != nil {
			return
		}

		result = append(result, r...)
		params = append(params, p...)
	}

	if len(wc.Values) > 1 || isSub {
		result = append(result, ")")
	}
	return
}

// placeholders returns n "?" joined by sep, e.g. placeholders(3, ",") == "?,?,?".
func placeholders(n int, sep string) string {
	if n <= 0 {
		return ""
	}
	return strings.TrimSuffix(strings.Repeat("?"+sep, n), sep)
}

func (b *Builder) withFields(query string) string {
	fields := "*"
	if len(b.ctx.Select) != 0 {
		fields = strings.Join(b.ctx.Select, ", ")
	}

	return strings.ReplaceAll(query, "{fields}", fields)
}

func (b *Builder) withCount(query, key string) string {
	if key == "" {
		key = "*"
	}
	return strings.ReplaceAll(query, "{fields}", fmt.Sprintf("count(%s)", key))
}

func (b *Builder) withLimitOffset(query string) string {
	if b.ctx.Limit != 0 {
		query += fmt.Sprintf(" LIMIT %d", b.ctx.Limit)

		if b.ctx.Offset != 0 {
			query += fmt.Sprintf(" OFFSET %d", b.ctx.Offset)
		}
	}

	return query
}

func (b *Builder) build(withOrderBy bool) (res string, params []any, err error) {
	var result []string

	if len(b.ctx.Table) == 0 {
		return res, params, errors.New("one table need to be defined")
	}

	result = append(result, "SELECT")

	if b.ctx.WithDistinct {
		result = append(result, "DISTINCT")
	}

	result = append(result, fmt.Sprintf(
		"%s FROM %s",
		"{fields}",
		strings.Join(b.ctx.Table, ", "),
	))

	for _, j := range b.ctx.Join {
		var r []string
		var p []any

		result = append(result, fmt.Sprintf("%s %s ON", j.Type, j.Table))

		r, p, err = buildWhere(j.Where, true, true, false)
		if err != nil {
			return
		}

		result = append(result, r...)
		params = append(params, p...)
	}

	for i, wc := range b.ctx.Where {
		var r []string
		var p []any

		r, p, err = buildWhere(wc, i == 0, false, false)
		if err != nil {
			return
		}

		result = append(result, r...)
		params = append(params, p...)
	}

	if len(b.ctx.GroupBy) != 0 {
		result = append(result, fmt.Sprintf("GROUP BY %s", strings.Join(b.ctx.GroupBy, ", ")))
	}

	if withOrderBy && len(b.ctx.OrderBy) != 0 {
		result = append(result, "ORDER BY")

		var ol []string
		for _, o := range b.ctx.OrderBy {
			ol = append(ol, fmt.Sprintf("%s %s", o.Key, o.Direction))
		}

		result = append(result, strings.Join(ol, ", "))
	}

	return strings.Join(result, " "), params, nil
}
