package melody

type Meta struct {
	Header    map[string]MetaField `json:"header"`
	Total     int                  `json:"total"`
	NbPerPage int                  `json:"nb_per_page"`
	Pages     int                  `json:"pages"`
}

type MetaField struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value any    `json:"value,omitempty"`
}

type Result struct {
	Meta Meta `json:"meta"`
	Body any  `json:"body"`
}

// NewResult wraps a body with pagination metadata. perPage and page mirror the
// values consumed by Builder.WithQueryParams; Pages is the total page count.
func NewResult(body any, total, perPage, page int) Result {
	pages := 0
	if perPage > 0 {
		pages = (total + perPage - 1) / perPage // ceil(total/perPage)
	}
	return Result{
		Meta: Meta{
			Total:     total,
			NbPerPage: perPage,
			Pages:     pages,
		},
		Body: body,
	}
}
