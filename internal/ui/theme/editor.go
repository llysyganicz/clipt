package theme

import (
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	// LineNumber styles the edit mode's line number gutter.
	LineNumber = lipgloss.NewStyle().Foreground(lipgloss.Color(gray))
	// CursorLineNumber styles the gutter on the cursor's line.
	CursorLineNumber = lipgloss.NewStyle().Foreground(lipgloss.Color(yellow))
)

// cursorColor is the color of the (virtual) cursor block in every input.
var cursorColor = lipgloss.Color(fg)

// EditorStyles returns the styles for the edit mode's text area: text in the
// normal foreground, no cursor line background and no prompt styling (the
// editor draws its own gutter through the prompt). Selection is left
// unstyled because the editor disables the selection keys.
func EditorStyles() textarea.Styles {
	text := lipgloss.NewStyle().Foreground(lipgloss.Color(fg))
	focused := textarea.StyleState{
		Base:             lipgloss.NewStyle(),
		CursorLine:       text,
		CursorLineNumber: CursorLineNumber,
		EndOfBuffer:      lipgloss.NewStyle(),
		LineNumber:       LineNumber,
		Placeholder:      Hint,
		Prompt:           lipgloss.NewStyle(),
		Text:             text,
	}
	blurred := focused
	blurred.CursorLineNumber = focused.LineNumber
	return textarea.Styles{
		Focused: focused,
		Blurred: blurred,
		Cursor:  textarea.CursorStyle{Color: cursorColor, Shape: tea.CursorBlock, Blink: true},
	}
}

// InputStyles returns the styles for every single-line text input: the
// prompt in the Prompt style, unstyled text in both focus states, hints for
// placeholder and suggestion text, and the same cursor as the editor.
func InputStyles() textinput.Styles {
	state := textinput.StyleState{
		Prompt:      Prompt,
		Text:        lipgloss.NewStyle(),
		Placeholder: Hint,
		Suggestion:  Hint,
	}
	return textinput.Styles{
		Focused: state,
		Blurred: state,
		Cursor:  textinput.CursorStyle{Color: cursorColor, Shape: tea.CursorBlock, Blink: true},
	}
}
