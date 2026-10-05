package theme

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
)

// EditorStyles returns the focused and blurred styles for the edit mode's
// text area: gray line numbers, a yellow number on the cursor line, text in
// the normal foreground and no cursor line background.
func EditorStyles() (focused, blurred textarea.Style) {
	text := lipgloss.NewStyle().Foreground(lipgloss.Color(fg))
	focused = textarea.Style{
		Base:             lipgloss.NewStyle(),
		CursorLine:       text,
		CursorLineNumber: lipgloss.NewStyle().Foreground(lipgloss.Color(yellow)),
		EndOfBuffer:      lipgloss.NewStyle(),
		LineNumber:       lipgloss.NewStyle().Foreground(lipgloss.Color(gray)),
		Placeholder:      Hint,
		Prompt:           lipgloss.NewStyle(),
		Text:             text,
	}
	blurred = focused
	blurred.CursorLineNumber = focused.LineNumber
	return focused, blurred
}
