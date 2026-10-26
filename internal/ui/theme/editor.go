package theme

import (
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// EditorStyles returns the styles for the edit mode's text area: text in the
// normal foreground, no cursor line background, no prompt styling (the
// editor draws its own gutter through the prompt) and the selection on the
// palette's selection background with the normal foreground.
func EditorStyles() textarea.Styles {
	text := lipgloss.NewStyle().Foreground(lipgloss.Color(current.Foreground))
	focused := textarea.StyleState{
		Base:             lipgloss.NewStyle(),
		CursorLine:       text,
		CursorLineNumber: CursorLineNumber,
		EndOfBuffer:      lipgloss.NewStyle(),
		LineNumber:       LineNumber,
		Placeholder:      Hint,
		Prompt:           lipgloss.NewStyle(),
		Selection:        lipgloss.NewStyle().Background(lipgloss.Color(current.Selection)),
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
// prompt in the Prompt style, typed text in the palette foreground in both
// focus states, hints for placeholder and suggestion text, and the same
// cursor as the editor.
func InputStyles() textinput.Styles {
	state := textinput.StyleState{
		Prompt:      Prompt,
		Text:        lipgloss.NewStyle().Foreground(lipgloss.Color(current.Foreground)),
		Placeholder: Hint,
		Suggestion:  Hint,
	}
	return textinput.Styles{
		Focused: state,
		Blurred: state,
		Cursor:  textinput.CursorStyle{Color: cursorColor, Shape: tea.CursorBlock, Blink: true},
	}
}
