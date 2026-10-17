package ui

import (
	"fmt"
	"os"

	"github.com/muesli/termenv"
)

var output = termenv.NewOutput(os.Stdout, termenv.WithProfile(termenv.ColorProfile()))

// Success formats text in green.
func Success(s string) string {
	return output.String(s).Foreground(output.Color("2")).String()
}

// Error formats text in red.
func Error(s string) string {
	return output.String(s).Foreground(output.Color("1")).String()
}

// Warn formats text in yellow.
func Warn(s string) string {
	return output.String(s).Foreground(output.Color("3")).String()
}

// Info formats text in blue.
func Info(s string) string {
	return output.String(s).Foreground(output.Color("4")).String()
}

// Dim formats text in a muted/dark color.
func Dim(s string) string {
	return output.String(s).Faint().String()
}

// Bold formats text in bold.
func Bold(s string) string {
	return output.String(s).Bold().String()
}

// Tag styles for repo and owner tags in the selectors. These are kept
// separate from the status colors above so tags can be restyled without
// affecting status output.

// PrivateTag formats a private repo tag in magenta.
func PrivateTag(s string) string {
	return output.String(s).Foreground(output.Color("5")).String()
}

// InternalTag formats an internal repo tag in cyan.
func InternalTag(s string) string {
	return output.String(s).Foreground(output.Color("6")).String()
}

// ForkTag formats a fork repo tag in yellow.
func ForkTag(s string) string {
	return output.String(s).Foreground(output.Color("3")).String()
}

// OrgTag formats an org owner tag in blue.
func OrgTag(s string) string {
	return output.String(s).Foreground(output.Color("4")).String()
}

// UserTag formats a user owner tag in magenta.
func UserTag(s string) string {
	return output.String(s).Foreground(output.Color("5")).String()
}

// Successf is a formatted version of Success.
func Successf(format string, a ...any) string {
	return Success(fmt.Sprintf(format, a...))
}

// Errorf is a formatted version of Error.
func Errorf(format string, a ...any) string {
	return Error(fmt.Sprintf(format, a...))
}

// Warnf is a formatted version of Warn.
func Warnf(format string, a ...any) string {
	return Warn(fmt.Sprintf(format, a...))
}
