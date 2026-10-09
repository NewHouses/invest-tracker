// Package money provides consistent currency formatting helpers used across
// reports. All amounts in this codebase are USD; the helpers prefix values
// with the $ symbol.
package money

import "fmt"

// USD formats a value as "$1500.00" (or "-$33.18" when negative, e.g. the net
// contribution of an asset that was sold for more than it cost).
func USD(v float64) string {
	if v < 0 {
		return fmt.Sprintf("-$%.2f", -v)
	}
	return fmt.Sprintf("$%.2f", v)
}

// SignedUSD formats a value as "+$150.00" or "-$50.00".
func SignedUSD(v float64) string {
	if v >= 0 {
		return fmt.Sprintf("+$%.2f", v)
	}
	return fmt.Sprintf("-$%.2f", -v)
}
