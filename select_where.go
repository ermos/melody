package melody

import "strings"

func (b *Builder) GroupWhere(sub SubBuilderFunc) *Builder {
	return b.sub(sub, false)
}

func (b *Builder) GroupOn(sub SubBuilderFunc) *Builder {
	return b.sub(sub, false)
}

func (b *Builder) OrGroupWhere(sub SubBuilderFunc) *Builder {
	return b.sub(sub, true)
}

func (b *Builder) OrGroupOn(sub SubBuilderFunc) *Builder {
	return b.sub(sub, true)
}

func (b *Builder) Where(key string, operator string, values ...any) *Builder {
	return b.where(key, operator, values, false, false)
}

func (b *Builder) OrWhere(key string, operator string, values ...any) *Builder {
	return b.where(key, operator, values, true, false)
}

// WhereRaw adds a raw boolean predicate, e.g.
// WhereRaw("LOWER(name) LIKE LOWER(?)", "%bob%") or
// WhereRaw("EXISTS (SELECT 1 FROM orders WHERE orders.user_id = users.id)").
func (b *Builder) WhereRaw(expr string, args ...any) *Builder {
	return b.raw(expr, false, args)
}

func (b *Builder) OrWhereRaw(expr string, args ...any) *Builder {
	return b.raw(expr, true, args)
}

func (b *Builder) WhereNull(key string) *Builder    { return b.raw(key+" IS NULL", false, nil) }
func (b *Builder) OrWhereNull(key string) *Builder  { return b.raw(key+" IS NULL", true, nil) }
func (b *Builder) WhereNotNull(key string) *Builder { return b.raw(key+" IS NOT NULL", false, nil) }
func (b *Builder) OrWhereNotNull(key string) *Builder {
	return b.raw(key+" IS NOT NULL", true, nil)
}

func (b *Builder) raw(expr string, isOr bool, args []any) *Builder {
	b.ctx.Where = append(b.ctx.Where, WhereContext{
		Values: []where{{Raw: expr, Values: args, IsOr: isOr}},
	})
	return b
}

func (b *Builder) On(firstKey string, operator string, secondKey string) *Builder {
	return b.where(firstKey, operator, []any{secondKey}, false, true)
}

func (b *Builder) OrOn(firstKey string, operator string, secondKey string) *Builder {
	return b.where(firstKey, operator, []any{secondKey}, true, true)
}

func (b *Builder) sub(sub SubBuilderFunc, isOr bool) *Builder {
	wc := &WhereContext{
		IsOr: isOr,
	}

	sub(wc)

	b.ctx.Where = append(b.ctx.Where, *wc)

	return b
}

func (b *Builder) where(key, operator string, values []any, isOr, isOn bool) *Builder {
	b.ctx.Where = append(b.ctx.Where, WhereContext{
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
	return b
}
