package output

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
)

var (
	ColorBlue   = lipgloss.Color("4")
	ColorGreen  = lipgloss.Color("2")
	ColorYellow = lipgloss.Color("3")
	ColorRed    = lipgloss.Color("1")
	ColorDim    = lipgloss.Color("8")

	infoMarker = lipgloss.NewStyle().Bold(true).Foreground(ColorBlue).Render("==>")
	bold       = lipgloss.NewStyle().Bold(true)

	warnTag = lipgloss.NewStyle().Foreground(ColorYellow).Render("warning:")
	errTag  = lipgloss.NewStyle().Foreground(ColorRed).Render("error:")

	Success = lipgloss.NewStyle().Foreground(ColorGreen)
	Stale   = lipgloss.NewStyle().Foreground(ColorRed)
	Skip    = lipgloss.NewStyle().Foreground(ColorYellow)
	Dim     = lipgloss.NewStyle().Foreground(ColorDim)
)

// Info prints "==> <msg>" with the marker in blue and the message bold.
func Info(msg string) {
	fmt.Fprintln(os.Stdout, infoMarker+" "+bold.Render(msg))
}

// Infof is like Info but with formatting.
func Infof(format string, a ...any) {
	Info(fmt.Sprintf(format, a...))
}

// Warn prints "warning: <msg>" to stderr.
func Warn(msg string) {
	fmt.Fprintln(os.Stderr, warnTag+" "+msg)
}

// Warnf is like Warn but with formatting.
func Warnf(format string, a ...any) {
	Warn(fmt.Sprintf(format, a...))
}

// Errorln prints "error: <msg>" to stderr.
func Errorln(msg string) {
	fmt.Fprintln(os.Stderr, errTag+" "+msg)
}

// Fprintf writes to w without styling.
func Fprintf(w io.Writer, format string, a ...any) {
	fmt.Fprintf(w, format, a...)
}

// Stdout returns os.Stdout — exported so callers don't have to import os
// just to print to it.
func Stdout() io.Writer { return os.Stdout }
