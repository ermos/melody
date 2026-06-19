package melody

// Melody is a builder factory pre-configured with a dialect, so you set the
// dialect once at startup instead of repeating .Dialect() on every query.
//
//	m := melody.With(melody.Postgres) // keep this around (e.g. a package var)
//	m.New("users").Where("id", "=", 1).Get()
//	m.NewInsert("users").Set("name", "bob").Get()
//
// The package-level New/NewInsert/... still work unchanged; and any builder
// returned here can override the dialect with a further .Dialect() call.
type Melody struct {
	dialect Dialect
}

// With returns a factory whose builders default to d.
func With(d Dialect) *Melody {
	return &Melody{dialect: d}
}

func (m *Melody) New(tables ...string) *Builder {
	return New(tables...).Dialect(m.dialect)
}

func (m *Melody) NewInsert(table string) *InsertBuilder {
	return NewInsert(table).Dialect(m.dialect)
}

func (m *Melody) NewUpdate(table string) *UpdateBuilder {
	return NewUpdate(table).Dialect(m.dialect)
}

func (m *Melody) NewDelete(table string) *DeleteBuilder {
	return NewDelete(table).Dialect(m.dialect)
}
