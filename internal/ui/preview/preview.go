// Package preview implements the reading mode: the current note rendered as
// markdown in a scrolling viewport, with edit, rename and delete actions.
package preview

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/llysyganicz/clipt/internal/notes"
	"github.com/llysyganicz/clipt/internal/ui/theme"
)

// BackMsg is emitted on Esc.
type BackMsg struct{}

// EditMsg is emitted when the user asks to edit the note.
type EditMsg struct{ Note notes.Note }

// RenamedMsg is emitted after the note file was renamed.
type RenamedMsg struct{ Note notes.Note }

// DeletedMsg is emitted after the note file was deleted.
type DeletedMsg struct{}

const (
	defaultWidth  = 80
	defaultHeight = 24
	hints         = "e edit · r rename · d delete · esc back · ctrl+q quit"
)

type state int

const (
	stateView state = iota
	stateRename
	stateConfirm
)

// Model is the preview Bubble Tea model.
type Model struct {
	store   notes.Store
	note    notes.Note
	content string
	width   int
	height  int
	vp      viewport.Model
	state   state
	input   textinput.Model
	err     string
}

// New reads the note and returns a preview of it. It fails when the note
// cannot be read so the caller can fall back to the finder.
func New(store notes.Store, note notes.Note) (Model, error) {
	content, err := store.Read(note.Name)
	if err != nil {
		return Model{}, err
	}
	vp := viewport.New(defaultWidth, defaultHeight-1)
	vp.KeyMap = viewport.KeyMap{
		Up:       key.NewBinding(key.WithKeys("up")),
		Down:     key.NewBinding(key.WithKeys("down")),
		PageUp:   key.NewBinding(key.WithKeys("pgup")),
		PageDown: key.NewBinding(key.WithKeys("pgdown")),
	}
	m := Model{
		store:   store,
		note:    note,
		content: content,
		width:   defaultWidth,
		height:  defaultHeight,
		vp:      vp,
	}
	m.render()
	return m, nil
}

// render re-renders the markdown at the current width and scrolls to the top.
func (m *Model) render() {
	m.vp.Width = m.width
	m.vp.Height = max(m.height-1, 0)
	out := m.content
	if r, err := glamour.NewTermRenderer(
		glamour.WithStyles(theme.Markdown),
		glamour.WithWordWrap(m.width),
	); err == nil {
		if s, err := r.Render(m.content); err == nil {
			out = s
		}
	}
	m.vp.SetContent(strings.TrimRight(out, "\n"))
	m.vp.GotoTop()
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.render()
		return m, nil
	case tea.KeyMsg:
		m.err = ""
		if msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyCtrlQ {
			return m, tea.Quit
		}
		switch m.state {
		case stateRename:
			return m.updateRename(msg)
		case stateConfirm:
			return m.updateConfirm(msg)
		}
		return m.updateView(msg)
	}
	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

func (m Model) updateView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.Type == tea.KeyEsc:
		return m, func() tea.Msg { return BackMsg{} }
	case msg.String() == "e":
		note := m.note
		return m, func() tea.Msg { return EditMsg{Note: note} }
	case msg.String() == "r":
		ti := textinput.New()
		ti.Prompt = "Rename: "
		ti.PromptStyle = theme.Prompt
		ti.SetValue(m.note.Name)
		ti.CursorEnd()
		ti.Focus()
		m.input = ti
		m.state = stateRename
		return m, textinput.Blink
	case msg.String() == "d":
		m.state = stateConfirm
		return m, nil
	}
	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

func (m Model) updateRename(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.state = stateView
		return m, nil
	case tea.KeyEnter:
		name := strings.TrimSpace(m.input.Value())
		if name == m.note.Name {
			m.state = stateView
			return m, nil
		}
		if err := m.store.Rename(m.note.Name, name); err != nil {
			m.err = err.Error()
			return m, nil
		}
		m.note.Name = name
		m.state = stateView
		note := m.note
		return m, func() tea.Msg { return RenamedMsg{Note: note} }
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.Type == tea.KeyEsc, msg.String() == "n", msg.String() == "N":
		m.state = stateView
	case msg.String() == "y", msg.String() == "Y":
		m.state = stateView
		if err := m.store.Delete(m.note.Name); err != nil {
			m.err = err.Error()
			return m, nil
		}
		return m, func() tea.Msg { return DeletedMsg{} }
	}
	return m, nil
}

// statusBar renders the bottom line for the current state.
func (m Model) statusBar() string {
	switch m.state {
	case stateRename:
		return m.input.View()
	case stateConfirm:
		return theme.Prompt.Render("Delete " + m.note.Name + "? (y/n)")
	}
	left := theme.Hint.Render(hints)
	avail := m.width - lipgloss.Width(left) - 1
	name := []rune(m.note.Name)
	if avail <= 0 {
		return left
	}
	if len(name) > avail {
		name = append([]rune{'…'}, name[len(name)-avail+1:]...)
	}
	return left + lipgloss.PlaceHorizontal(m.width-lipgloss.Width(left), lipgloss.Right, theme.Prompt.Render(string(name)))
}

// View implements tea.Model.
func (m Model) View() string {
	body := m.vp.View()
	if m.err != "" {
		lines := strings.Split(body, "\n")
		lines[len(lines)-1] = theme.Error.Render(m.err)
		body = strings.Join(lines, "\n")
	}
	return body + "\n" + m.statusBar()
}
