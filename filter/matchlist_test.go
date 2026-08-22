package filter

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestMatchlist(t *testing.T) {
	cases := map[string]struct {
		list     []string
		presence bool
		match    string
		res      bool
	}{
		"positive match":                     {[]string{"a", "b"}, true, "a", true},
		"positive match 2":                   {[]string{"a", "b"}, true, "b", true},
		"positive no match":                  {[]string{"a", "b"}, true, "c", false},
		"positive no match case insensitive": {[]string{"a", "b"}, true, "A", false},
		"negative match":                     {[]string{"a", "b"}, false, "a", false},
		"negative match 2":                   {[]string{"a", "b"}, false, "b", false},
		"negative no match":                  {[]string{"a", "b"}, false, "c", true},
		"negative no match case insensitive": {[]string{"a", "b"}, false, "A", true},

		"positive empty list": {[]string{}, true, "a", false},
		"negative empty list": {[]string{}, false, "a", true},
	}
	for n, c := range cases {
		t.Run(n, func(t *testing.T) {
			t.Log(c.list, c.match, c.presence, c.res)
			m := newMatchlist(c.list, c.presence)
			r := m.match(c.match)
			if r != c.res {
				t.Error("wrong result")
			}
		})
	}
}

func BenchmarkMatchlist(b *testing.B) {
	for _, i := range []int{0, 1, 2, 3, 4, 5, 6, 8, 10, 15, 20, 25, 30, 40, 50, 75, 100, 300, 1000} {
		var s []string
		for j := 0; j < i; j++ {
			s = append(s, fmt.Sprintf("%30d", j))
		}
		m := newMatchlist(s, true)
		c := strings.Repeat(" ", 30)
		b.Run(strconv.Itoa(i), func(b *testing.B) {
			for j := 0; j < b.N; j++ {
				_ = m.match(c)
			}
		})
	}
}
