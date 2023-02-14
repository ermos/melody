package melody

type join struct {
	Type  string
	Table string
	Where WhereContext
}

func (b *Builder) Join(table string, sub SubBuilderFunc) *Builder {
	return b.join(Join, table, sub)
}

func (b *Builder) InnerJoin(table string, sub SubBuilderFunc) *Builder {
	return b.join(InnerJoin, table, sub)
}

func (b *Builder) CrossJoin(table string, sub SubBuilderFunc) *Builder {
	return b.join(CrossJoin, table, sub)
}

func (b *Builder) LeftJoin(table string, sub SubBuilderFunc) *Builder {
	return b.join(LeftJoin, table, sub)
}

func (b *Builder) RightJoin(table string, sub SubBuilderFunc) *Builder {
	return b.join(RightJoin, table, sub)
}

func (b *Builder) FullJoin(table string, sub SubBuilderFunc) *Builder {
	return b.join(FullJoin, table, sub)
}

func (b *Builder) SelfJoin(table string, sub SubBuilderFunc) *Builder {
	return b.join(SelfJoin, table, sub)
}

func (b *Builder) NaturalJoin(table string, sub SubBuilderFunc) *Builder {
	return b.join(NaturalJoin, table, sub)
}

func (b *Builder) UnionJoin(table string, sub SubBuilderFunc) *Builder {
	return b.join(UnionJoin, table, sub)
}

func (b *Builder) join(joinType string, table string, sub SubBuilderFunc) *Builder {
	wc := &WhereContext{}
	sub(wc)

	b.ctx.Join = append(b.ctx.Join, join{
		Type:  joinType,
		Table: table,
		Where: *wc,
	})

	return b
}
