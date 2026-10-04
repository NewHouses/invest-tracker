// Package colors provides ANSI escape codes for terminal background colors
// used to highlight rows in historical reports.
package colors

const (
	// Reset returns the terminal to default colors.
	Reset = "\x1b[0m"

	// GreenBG is a pale green background with black text — for positive G/P.
	GreenBG = "\x1b[48;5;157;30m"

	// RedBG is a pale red background with black text — for negative G/P.
	RedBG = "\x1b[48;5;217;30m"

	// YellowBG is a pale yellow background with black text — for zero G/P.
	YellowBG = "\x1b[48;5;229;30m"
)

// ForGain returns the background color escape for a gain value. Callers must
// only invoke this when the row has computable metrics (holding > 0).
func ForGain(gain float64) string {
	switch {
	case gain > 0:
		return GreenBG
	case gain < 0:
		return RedBG
	default:
		return YellowBG
	}
}
