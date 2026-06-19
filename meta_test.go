package melody

import "testing"

func TestNewResultPages(t *testing.T) {
	cases := []struct {
		total, perPage, wantPages int
	}{
		{0, 10, 0},
		{10, 10, 1},
		{11, 10, 2},
		{25, 10, 3},
		{5, 0, 0}, // perPage 0 -> no division
	}
	for _, c := range cases {
		r := NewResult(nil, c.total, c.perPage, 1)
		if r.Meta.Pages != c.wantPages {
			t.Errorf("NewResult(total=%d,perPage=%d).Pages = %d, want %d",
				c.total, c.perPage, r.Meta.Pages, c.wantPages)
		}
	}
}
