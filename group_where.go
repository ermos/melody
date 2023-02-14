package melody

import "strings"

func (b *WhereContext) GroupWhere(sub SubBuilderFunc) *WhereContext {
	return b.sub(sub, false)
}

func (b *WhereContext) GroupOn(sub SubBuilderFunc) *WhereContext {
	return b.sub(sub, false)
}

func (b *WhereContext) OrGroupWhere(sub SubBuilderFunc) *WhereContext {
	return b.sub(sub, false)
}

func (b *WhereContext) OrGroupOn(sub SubBuilderFunc) *WhereContext {
	return b.sub(sub, false)
}

func (b *WhereContext) Where(key string, operator string, values ...interface{}) *WhereContext {
	return b.where(key, operator, values, false, false)
}

func (b *WhereContext) OrWhere(key string, operator string, values ...interface{}) *WhereContext {
	return b.where(key, operator, values, true, false)
}

func (b *WhereContext) On(firstKey string, operator string, secondKey string) *WhereContext {
	return b.where(firstKey, operator, []interface{}{secondKey}, false, true)
}

func (b *WhereContext) OrOn(firstKey string, operator string, secondKey string) *WhereContext {
	return b.where(firstKey, operator, []interface{}{secondKey}, true, true)
}

func (b *WhereContext) sub(sub SubBuilderFunc, isOr bool) *WhereContext {
	wc := &WhereContext{
		IsOr: isOr,
	}

	sub(wc)

	b.Sub = append(b.Sub, *wc)

	return b
}

func (b *WhereContext) where(key, operator string, values []interface{}, isOr, isOn bool) *WhereContext {
	b.Values = append(b.Values, where{
		Key:      key,
		Operator: strings.ToUpper(operator),
		Values:   values,
		IsOr:     isOr,
		IsOn:     isOn,
	})
	return b
}
