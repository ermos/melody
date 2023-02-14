package melody

import "strings"

func (u *UpdateBuilder) GroupWhere(sub SubBuilderFunc) *UpdateBuilder {
	return u.sub(sub, false)
}

func (u *UpdateBuilder) GroupOn(sub SubBuilderFunc) *UpdateBuilder {
	return u.sub(sub, false)
}

func (u *UpdateBuilder) OrGroupWhere(sub SubBuilderFunc) *UpdateBuilder {
	return u.sub(sub, false)
}

func (u *UpdateBuilder) OrGroupOn(sub SubBuilderFunc) *UpdateBuilder {
	return u.sub(sub, false)
}

func (u *UpdateBuilder) Where(key string, operator string, values ...interface{}) *UpdateBuilder {
	return u.where(key, operator, values, false, false)
}

func (u *UpdateBuilder) OrWhere(key string, operator string, values ...interface{}) *UpdateBuilder {
	return u.where(key, operator, values, true, false)
}

func (u *UpdateBuilder) On(firstKey string, operator string, secondKey string) *UpdateBuilder {
	return u.where(firstKey, operator, []interface{}{secondKey}, false, true)
}

func (u *UpdateBuilder) OrOn(firstKey string, operator string, secondKey string) *UpdateBuilder {
	return u.where(firstKey, operator, []interface{}{secondKey}, true, true)
}

func (u *UpdateBuilder) sub(sub SubBuilderFunc, isOr bool) *UpdateBuilder {
	wc := &WhereContext{
		IsOr: isOr,
	}

	sub(wc)

	u.ctx.Where = append(u.ctx.Where, *wc)

	return u
}

func (u *UpdateBuilder) where(key, operator string, values []interface{}, isOr, isOn bool) *UpdateBuilder {
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
