// Package ui owns terminal styling and shared interactive helpers.
//
// The Python tool uses raw ANSI escape codes via a tiny Colors helper. We use
// lipgloss for the same effect but with adaptive colour, automatic NO_COLOR
// support, and proper width calculations for two-column layouts.
package ui

import "github.com/charmbracelet/lipgloss"

var (
	Title    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	Heading  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	Subtle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	Success  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	Warn     = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	ErrorSty = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	Accent   = lipgloss.NewStyle().Foreground(lipgloss.Color("13"))
	Favorite = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	Muted    = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
)

// Iconf prefixes a string with an indicator glyph styled in the given style.
func Iconf(style lipgloss.Style, glyph, msg string) string {
	return style.Render(glyph) + " " + msg
}

func Successf(msg string) string { return Iconf(Success, "[ok]", msg) }
func Warnf(msg string) string    { return Iconf(Warn, "[!]", msg) }
func Errorf(msg string) string   { return Iconf(ErrorSty, "[x]", msg) }
func Infof(msg string) string    { return Iconf(Accent, "[i]", msg) }
