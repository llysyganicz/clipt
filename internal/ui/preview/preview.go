// Package preview implements the reading mode: the current note rendered as
// markdown in a scrolling viewport, with edit, rename and delete actions.
package preview

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"

	"github.com/llysyganicz/clipt/internal/frontmatter"
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
	header  string // tags header, or an invalid front matter error; may be empty
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
	vp := viewport.New(viewport.WithWidth(defaultWidth), viewport.WithHeight(defaultHeight-1))
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

// render re-renders the tags header and the markdown body at the current
// width and scrolls to the top. The front matter is never rendered as
// markdown; when it is not closed the whole file is the body.
func (m *Model) render() {
	fm, body, err := frontmatter.Split(m.content)
	var tags []string
	if err == nil && fm.Present {
		tags, err = frontmatter.Tags(fm.YAML())
	}
	switch {
	case err != nil:
		msg, _, _ := strings.Cut(err.Error(), "\n")
		m.header = theme.Error.Width(m.width).Render("invalid front matter: " + msg)
	case len(tags) > 0:
		m.header = theme.Tag.Width(m.width).Render(frontmatter.Display(tags))
	default:
		m.header = ""
	}
	headerLines := 0
	if m.header != "" {
		headerLines = lipgloss.Height(m.header)
	}

	m.vp.SetWidth(m.width)
	m.vp.SetHeight(max(m.height-1-headerLines, 0))
	out := body
	if r, err := glamour.NewTermRenderer(
		glamour.WithStyles(theme.Markdown),
		glamour.WithWordWrap(m.width),
	); err == nil {
		if s, err := r.Render(body); err == nil {
			out = s
		}
	}
	m.vp.SetContent(strings.TrimRight(out, "\n"))
	m.vp.GotoTop()
}

// Init returns no command.
func (m Model) Init() tea.Cmd { return nil }

// Update handles a message and returns the new model.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.render()
		return m, nil
	case tea.KeyPressMsg:
		m.err = ""
		if k := msg.String(); k == "ctrl+c" || k == "ctrl+q" {
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

func (m Model) updateView(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m, func() tea.Msg { return BackMsg{} }
	case "e":
		note := m.note
		return m, func() tea.Msg { return EditMsg{Note: note} }
	case "r":
		ti := textinput.New()
		ti.Prompt = "Rename: "
		ti.SetStyles(theme.InputStyles())
		ti.SetValue(m.note.Name)
		ti.CursorEnd()
		ti.Focus()
		m.input = ti
		m.state = stateRename
		return m, textinput.Blink
	case "d":
		m.state = stateConfirm
		return m, nil
	}
	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

func (m Model) updateRename(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = stateView
		return m, nil
	case "enter":
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

func (m Model) updateConfirm(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "n", "N":
		m.state = stateView
	case "y", "Y":
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

// View renders the preview.
func (m Model) View() string {
	body := m.vp.View()
	if m.err != "" {
		lines := strings.Split(body, "\n")
		lines[len(lines)-1] = theme.Error.Render(m.err)
		body = strings.Join(lines, "\n")
	}
	if m.header != "" {
		body = m.header + "\n" + body
	}
	return body + "\n" + m.statusBar()
}
