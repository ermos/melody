package melody

type Meta struct {
	Header    map[string]MetaField `json:"header"`
	Total     int                  `json:"total"`
	NbPerPage int                  `json:"nb_per_page"`
	Pages     int                  `json:"pages"`
}

type MetaField struct {
	Name  string      `json:"name"`
	Type  string      `json:"type"`
	Value interface{} `json:"value,omitempty"`
}

type Result struct {
	Meta Meta        `json:"meta"`
	Body interface{} `json:"body"`
}
