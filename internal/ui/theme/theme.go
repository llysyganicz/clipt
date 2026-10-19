// Package theme holds every color and named style used by the UI. Mode
// packages must use these styles and never reference color literals.
package theme

import "charm.land/lipgloss/v2"

// Gruvbox dark palette.
const (
	bg     = "#282828"
	fg     = "#ebdbb2"
	yellow = "#fabd2f"
	aqua   = "#8ec07c"
	red    = "#fb4934"
	gray   = "#928374"
)

var (
	// Normal is the default text style.
	Normal = lipgloss.NewStyle().Foreground(lipgloss.Color(fg)).Background(lipgloss.Color(bg))
	// Selected styles the highlighted row.
	Selected = lipgloss.NewStyle().Foreground(lipgloss.Color(yellow)).Bold(true)
	// Create styles the "Create <query>" row.
	Create = lipgloss.NewStyle().Foreground(lipgloss.Color(aqua))
	// Hint styles secondary, explanatory text.
	Hint = lipgloss.NewStyle().Foreground(lipgloss.Color(gray))
	// Error styles inline error lines.
	Error = lipgloss.NewStyle().Foreground(lipgloss.Color(red))
	// Prompt styles input prompts and confirmation questions.
	Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color(yellow))
	// Tag styles a note's tags in the preview header.
	Tag = lipgloss.NewStyle().Foreground(lipgloss.Color(aqua))
	// RowTag styles the dimmed tags after a name on finder and picker rows.
	RowTag = lipgloss.NewStyle().Foreground(lipgloss.Color(gray)).Faint(true)
)
