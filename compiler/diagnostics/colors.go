package diagnostics

import (
	"os"
	"strings"
)

// ANSI color codes for diagnostic rendering.
const (
	ColorRed    = "\033[31m"
	ColorYellow = "\033[33m"
	ColorCyan   = "\033[36m"
	ColorBold   = "\033[1m"
	ColorReset  = "\033[0m"
	ColorDim    = "\033[2m"
)

// ColorSupport detects whether the terminal supports ANSI colors.
// Returns false if NO_COLOR env var is set (any value), or TERM=dumb.
// Follows the no-color.org standard.
func ColorSupport() bool {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	if strings.EqualFold(os.Getenv("TERM"), "dumb") {
		return false
	}
	return true
}

// colorize wraps text with an ANSI color code if useColor is true.
func colorize(text, color string, useColor bool) string {
	if !useColor {
		return text
	}
	return color + text + ColorReset
}
