package pg

import (
	"strings"

	"github.com/Deimvis/go-ext/go1.25/xcheck/xmust"
)

// TsQuery represents Postgresql
// input for to_tsquery function
// (see https://www.postgresql.org/docs/current/textsearch-controls.html#TEXTSEARCH-PARSING-QUERIES)
type TsQuery string

func (q *TsQuery) Raw() string {
	return string(*q)
}

func (q *TsQuery) Escaped() string {
	s := string(*q)
	b := strings.Builder{}
	for _, c := range s {
		if c == '\\' {
			xmust.NoErr(b.WriteByte('\\'))
		}
		_ = xmust.Do(b.WriteRune(c))
	}
	return b.String()
}

func (q *TsQuery) String() string {
	if q == nil {
		return "<nil>"
	}
	return string(*q)
}
