package melody

import "strings"

func (b *Builder) GroupWhere(sub SubBuilderFunc) *Builder {
	return b.sub(sub, false)
}

func (b *Builder) GroupOn(sub SubBuilderFunc) *Builder {
	return b.sub(sub, false)
}

func (b *Builder) OrGroupWhere(sub SubBuilderFunc) *Builder {
	return b.sub(sub, false)
}

func (b *Builder) OrGroupOn(sub SubBuilderFunc) *Builder {
	return b.sub(sub, false)
}

func (b *Builder) Where(key string, operator string, values ...interface{}) *Builder {
	return b.where(key, operator, values, false, false)
}

func (b *Builder) OrWhere(key string, operator string, values ...interface{}) *Builder {
	return b.where(key, operator, values, true, false)
}

func (b *Builder) On(firstKey string, operator string, secondKey string) *Builder {
	return b.where(firstKey, operator, []interface{}{secondKey}, false, true)
}

func (b *Builder) OrOn(firstKey string, operator string, secondKey string) *Builder {
	return b.where(firstKey, operator, []interface{}{secondKey}, true, true)
}

func (b *Builder) sub(sub SubBuilderFunc, isOr bool) *Builder {
	wc := &WhereContext{
		IsOr: isOr,
	}

	sub(wc)

	b.ctx.Where = append(b.ctx.Where, *wc)

	return b
}

func (b *Builder) where(key, operator string, values []interface{}, isOr, isOn bool) *Builder {
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
