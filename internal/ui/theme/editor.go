package theme

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// LineNumber styles the edit mode's line number gutter.
	LineNumber = lipgloss.NewStyle().Foreground(lipgloss.Color(gray))
	// CursorLineNumber styles the gutter on the cursor's line.
	CursorLineNumber = lipgloss.NewStyle().Foreground(lipgloss.Color(yellow))
)

// EditorStyles returns the focused and blurred styles for the edit mode's
// text area: text in the normal foreground, no cursor line background and
// no prompt styling (the editor draws its own gutter through the prompt).
func EditorStyles() (focused, blurred textarea.Style) {
	text := lipgloss.NewStyle().Foreground(lipgloss.Color(fg))
	focused = textarea.Style{
		Base:             lipgloss.NewStyle(),
		CursorLine:       text,
		CursorLineNumber: CursorLineNumber,
		EndOfBuffer:      lipgloss.NewStyle(),
		LineNumber:       LineNumber,
		Placeholder:      Hint,
		Prompt:           lipgloss.NewStyle(),
		Text:             text,
	}
	blurred = focused
	blurred.CursorLineNumber = focused.LineNumber
	return focused, blurred
}
