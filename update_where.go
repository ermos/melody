package melody

import "strings"

func (u *UpdateBuilder) GroupWhere(sub SubBuilderFunc) *UpdateBuilder {
	return u.sub(sub, false)
}

func (u *UpdateBuilder) GroupOn(sub SubBuilderFunc) *UpdateBuilder {
	return u.sub(sub, false)
}

func (u *UpdateBuilder) OrGroupWhere(sub SubBuilderFunc) *UpdateBuilder {
	return u.sub(sub, true)
}

func (u *UpdateBuilder) OrGroupOn(sub SubBuilderFunc) *UpdateBuilder {
	return u.sub(sub, true)
}

func (u *UpdateBuilder) Where(key string, operator string, values ...any) *UpdateBuilder {
	return u.where(key, operator, values, false, false)
}

func (u *UpdateBuilder) OrWhere(key string, operator string, values ...any) *UpdateBuilder {
	return u.where(key, operator, values, true, false)
}

func (u *UpdateBuilder) WhereRaw(expr string, args ...any) *UpdateBuilder {
	return u.raw(expr, false, args)
}

func (u *UpdateBuilder) OrWhereRaw(expr string, args ...any) *UpdateBuilder {
	return u.raw(expr, true, args)
}

func (u *UpdateBuilder) WhereNull(key string) *UpdateBuilder {
	return u.raw(key+" IS NULL", false, nil)
}
func (u *UpdateBuilder) OrWhereNull(key string) *UpdateBuilder {
	return u.raw(key+" IS NULL", true, nil)
}
func (u *UpdateBuilder) WhereNotNull(key string) *UpdateBuilder {
	return u.raw(key+" IS NOT NULL", false, nil)
}
func (u *UpdateBuilder) OrWhereNotNull(key string) *UpdateBuilder {
	return u.raw(key+" IS NOT NULL", true, nil)
}

func (u *UpdateBuilder) raw(expr string, isOr bool, args []any) *UpdateBuilder {
	u.ctx.Where = append(u.ctx.Where, WhereContext{
		Values: []where{{Raw: expr, Values: args, IsOr: isOr}},
	})
	return u
}

func (u *UpdateBuilder) On(firstKey string, operator string, secondKey string) *UpdateBuilder {
	return u.where(firstKey, operator, []any{secondKey}, false, true)
}

func (u *UpdateBuilder) OrOn(firstKey string, operator string, secondKey string) *UpdateBuilder {
	return u.where(firstKey, operator, []any{secondKey}, true, true)
}

func (u *UpdateBuilder) sub(sub SubBuilderFunc, isOr bool) *UpdateBuilder {
	wc := &WhereContext{
		IsOr: isOr,
	}

	sub(wc)

	u.ctx.Where = append(u.ctx.Where, *wc)

	return u
}

func (u *UpdateBuilder) where(key, operator string, values []any, isOr, isOn bool) *UpdateBuilder {
	u.ctx.Where = append(u.ctx.Where, WhereContext{
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
	return u
}
