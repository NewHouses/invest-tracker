package money_test

import (
	"testing"

	"invest-tracker/internal/money"
)

func TestUSD(t *testing.T) {
	cases := map[float64]string{
		1500:   "$1500.00",
		0:      "$0.00",
		-33.18: "-$33.18",
	}
	for v, want := range cases {
		if got := money.USD(v); got != want {
			t.Errorf("USD(%v) = %q, esperabamos %q", v, got, want)
		}
	}
}

func TestSignedUSD(t *testing.T) {
	cases := map[float64]string{
		150: "+$150.00",
		0:   "+$0.00",
		-50: "-$50.00",
	}
	for v, want := range cases {
		if got := money.SignedUSD(v); got != want {
			t.Errorf("SignedUSD(%v) = %q, esperabamos %q", v, got, want)
		}
	}
}
