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

// Bold formats text in bold.
func Bold(s string) string {
	return output.String(s).Bold().String()
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
