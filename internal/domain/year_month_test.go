package domain_test

import (
	"testing"

	"invest-tracker/internal/domain"
)

func TestAddMonths(t *testing.T) {
	cases := []struct {
		from domain.YearMonth
		n    int
		want domain.YearMonth
	}{
		{domain.YearMonth{Year: 2026, Month: 1}, 0, domain.YearMonth{Year: 2026, Month: 1}},
		{domain.YearMonth{Year: 2026, Month: 1}, 11, domain.YearMonth{Year: 2026, Month: 12}},
		{domain.YearMonth{Year: 2026, Month: 1}, 12, domain.YearMonth{Year: 2027, Month: 1}},
		{domain.YearMonth{Year: 2026, Month: 12}, 1, domain.YearMonth{Year: 2027, Month: 1}},
		{domain.YearMonth{Year: 2026, Month: 9}, 239, domain.YearMonth{Year: 2046, Month: 8}},
		{domain.YearMonth{Year: 2026, Month: 1}, -1, domain.YearMonth{Year: 2025, Month: 12}},
		{domain.YearMonth{Year: 2026, Month: 3}, -14, domain.YearMonth{Year: 2025, Month: 1}},
	}
	for _, c := range cases {
		if got := c.from.AddMonths(c.n); got != c.want {
			t.Errorf("%+v.AddMonths(%d) = %+v, esperabamos %+v", c.from, c.n, got, c.want)
		}
	}
}
