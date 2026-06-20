package melody

import "strings"

func (b *WhereContext) GroupWhere(sub SubBuilderFunc) *WhereContext {
	return b.sub(sub, false)
}

func (b *WhereContext) GroupOn(sub SubBuilderFunc) *WhereContext {
	return b.sub(sub, false)
}

func (b *WhereContext) OrGroupWhere(sub SubBuilderFunc) *WhereContext {
	return b.sub(sub, true)
}

func (b *WhereContext) OrGroupOn(sub SubBuilderFunc) *WhereContext {
	return b.sub(sub, true)
}

func (b *WhereContext) Where(key string, operator string, values ...any) *WhereContext {
	return b.where(key, operator, values, false, false)
}

func (b *WhereContext) OrWhere(key string, operator string, values ...any) *WhereContext {
	return b.where(key, operator, values, true, false)
}

func (b *WhereContext) WhereRaw(expr string, args ...any) *WhereContext {
	return b.raw(expr, false, args)
}

func (b *WhereContext) OrWhereRaw(expr string, args ...any) *WhereContext {
	return b.raw(expr, true, args)
}

func (b *WhereContext) WhereNull(key string) *WhereContext   { return b.raw(key+" IS NULL", false, nil) }
func (b *WhereContext) OrWhereNull(key string) *WhereContext { return b.raw(key+" IS NULL", true, nil) }
func (b *WhereContext) WhereNotNull(key string) *WhereContext {
	return b.raw(key+" IS NOT NULL", false, nil)
}
func (b *WhereContext) OrWhereNotNull(key string) *WhereContext {
	return b.raw(key+" IS NOT NULL", true, nil)
}

func (b *WhereContext) raw(expr string, isOr bool, args []any) *WhereContext {
	b.Values = append(b.Values, where{Raw: expr, Values: args, IsOr: isOr})
	return b
}

func (b *WhereContext) On(firstKey string, operator string, secondKey string) *WhereContext {
	return b.where(firstKey, operator, []any{secondKey}, false, true)
}

func (b *WhereContext) OrOn(firstKey string, operator string, secondKey string) *WhereContext {
	return b.where(firstKey, operator, []any{secondKey}, true, true)
}

func (b *WhereContext) sub(sub SubBuilderFunc, isOr bool) *WhereContext {
	wc := &WhereContext{
		IsOr: isOr,
	}

	sub(wc)

	b.Sub = append(b.Sub, *wc)

	return b
}

func (b *WhereContext) where(key, operator string, values []any, isOr, isOn bool) *WhereContext {
	b.Values = append(b.Values, where{
		Key:      key,
		Operator: strings.ToUpper(operator),
		Values:   values,
		IsOr:     isOr,
		IsOn:     isOn,
	})
	return b
}
