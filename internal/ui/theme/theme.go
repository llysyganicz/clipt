// Package theme holds every color and named style used by the UI. Mode
// packages must use these styles and never reference color literals.
//
// The styles are derived from a Palette by Apply. The package starts out
// with the built-in gruvbox dark theme; main applies the configured theme
// before any mode model is constructed.
package theme

import (
	"image/color"

	"charm.land/glamour/v2/ansi"
	"charm.land/lipgloss/v2"
)

var (
	// Normal is the default text style.
	Normal lipgloss.Style
	// Selected styles the highlighted row.
	Selected lipgloss.Style
	// Create styles the "Create <query>" row.
	Create lipgloss.Style
	// Hint styles secondary, explanatory text.
	Hint lipgloss.Style
	// Error styles inline error lines.
	Error lipgloss.Style
	// Prompt styles input prompts and confirmation questions.
	Prompt lipgloss.Style
	// Tag styles a note's tags in the preview header.
	Tag lipgloss.Style
	// RowTag styles the dimmed tags after a name on finder and picker rows.
	RowTag lipgloss.Style
	// LineNumber styles the edit mode's line number gutter.
	LineNumber lipgloss.Style
	// CursorLineNumber styles the gutter on the cursor's line.
	CursorLineNumber lipgloss.Style
	// Markdown styles rendered markdown in preview.
	Markdown ansi.StyleConfig
)

// current is the palette the styles were last built from.
var current Palette

// cursorColor is the color of the (virtual) cursor block in every input.
var cursorColor color.Color

func init() { Apply(mustBuiltin(Default)) }

// Apply rebuilds every style from p. Models capture styles when they are
// constructed, so call Apply before building any of them.
func Apply(p Palette) {
	current = p
	fg := lipgloss.Color(p.Foreground)
	Normal = lipgloss.NewStyle().Foreground(fg).Background(lipgloss.Color(p.Background))
	Selected = lipgloss.NewStyle().Foreground(lipgloss.Color(p.BrightYellow)).Bold(true)
	Create = lipgloss.NewStyle().Foreground(lipgloss.Color(p.BrightCyan))
	Hint = lipgloss.NewStyle().Foreground(lipgloss.Color(p.BrightBlack))
	Error = lipgloss.NewStyle().Foreground(lipgloss.Color(p.BrightRed))
	Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color(p.BrightYellow))
	Tag = lipgloss.NewStyle().Foreground(lipgloss.Color(p.BrightCyan))
	RowTag = lipgloss.NewStyle().Foreground(lipgloss.Color(p.BrightBlack)).Faint(true)
	LineNumber = lipgloss.NewStyle().Foreground(lipgloss.Color(p.BrightBlack))
	CursorLineNumber = lipgloss.NewStyle().Foreground(lipgloss.Color(p.BrightYellow))
	cursorColor = fg
	Markdown = markdownStyle(p)
}
