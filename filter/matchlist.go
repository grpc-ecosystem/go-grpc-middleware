package filter

type matchlist struct {
	m map[string]struct{}
	p bool
}

func newMatchlist(s []string, matchPresence bool) *matchlist {
	m := make(map[string]struct{}, len(s))
	for _, e := range s {
		m[e] = struct{}{}
	}
	return &matchlist{m, matchPresence}
}

func (m *matchlist) match(s string) bool {
	_, found := m.m[s]
	return found == m.p
}
