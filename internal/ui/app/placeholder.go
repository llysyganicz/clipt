package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/llysyganicz/clipt/internal/ui/theme"
)

// noteBackMsg is emitted by the edit placeholder on Esc.
type noteBackMsg struct{}

// placeholder stands in for edit until its change lands. It shows the note
// name and returns to preview on Esc.
type placeholder struct{ name string }

func (p placeholder) Init() tea.Cmd { return nil }

func (p placeholder) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.Type {
		case tea.KeyCtrlC:
			return p, tea.Quit
		case tea.KeyEsc:
			return p, func() tea.Msg { return noteBackMsg{} }
		}
	}
	return p, nil
}

func (p placeholder) View() string {
	return theme.Prompt.Render(p.name+" (edit)") + "\n" + theme.Hint.Render("esc back")
}
