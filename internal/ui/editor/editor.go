// Package editor implements the edit mode: the current note's raw markdown
// in an insert-only text area. Esc saves the note when it changed and hands
// control back to the caller.
package editor

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/llysyganicz/clipt/internal/notes"
	"github.com/llysyganicz/clipt/internal/ui/theme"
)

// DoneMsg is emitted on Esc once the note is saved (or was unchanged).
type DoneMsg struct{}

const (
	defaultWidth  = 80
	defaultHeight = 24
	hints         = "esc save & back"
)

// repositionMsg is a no-op message sent through the text area so it scrolls
// its viewport to the cursor.
type repositionMsg struct{}

// Model is the editor Bubble Tea model.
type Model struct {
	store    notes.Store
	note     notes.Note
	ta       textarea.Model
	baseline string
	width    int
	height   int
	err      string
}

// New reads the note and returns an editor on it with the cursor at the end.
// It fails when the note cannot be read so the caller can fall back to the
// finder.
func New(store notes.Store, note notes.Note) (Model, error) {
	content, err := store.Read(note.Name)
	if err != nil {
		return Model{}, err
	}
	content = strings.ReplaceAll(content, "\r\n", "\n")
	// The on-disk form ends with one newline; without dropping it the text
	// area would show an extra empty last line and put the cursor there.
	content = strings.TrimSuffix(content, "\n")

	ta := textarea.New()
	ta.ShowLineNumbers = true
	ta.Prompt = ""
	ta.CharLimit = 0
	ta.MaxHeight = 0
	ta.MaxWidth = 0
	ta.KeyMap = textarea.DefaultKeyMap
	ta.FocusedStyle, ta.BlurredStyle = theme.EditorStyles()
	ta.Focus()
	ta.SetValue(content)

	m := Model{
		store:    store,
		note:     note,
		ta:       ta,
		baseline: notes.Normalize(ta.Value()),
	}
	m.resize(defaultWidth, defaultHeight)
	return m, nil
}

// resize fits the text area above the status bar and scrolls it to the
// cursor.
func (m *Model) resize(width, height int) {
	m.width, m.height = width, height
	m.ta.SetWidth(width)
	m.ta.SetHeight(height - 1)
	m.scrollToCursor()
}

// scrollToCursor makes the text area scroll its viewport to the cursor. It
// only does that in Update, and only once a View call has filled the
// viewport.
func (m *Model) scrollToCursor() {
	_ = m.ta.View()
	m.ta, _ = m.ta.Update(repositionMsg{})
}

// reflow handles a terminal resize. The text area keeps its scroll offset
// across size changes, which leaves the view mostly blank after the terminal
// grows while scrolled down; reloading the buffer resets the offset, and the
// cursor is put back where it was before scrolling to it.
func (m *Model) reflow(width, height int) {
	row := m.ta.Line()
	info := m.ta.LineInfo()
	col := info.StartColumn + info.ColumnOffset
	m.ta.SetValue(m.ta.Value())
	for m.ta.Line() > row {
		m.ta.CursorUp()
	}
	m.ta.SetCursor(col)
	m.resize(width, height)
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return textarea.Blink }

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.reflow(msg.Width, msg.Height)
		return m, nil
	case tea.KeyMsg:
		m.err = ""
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEsc:
			return m.save()
		}
	}
	var cmd tea.Cmd
	m.ta, cmd = m.ta.Update(msg)
	return m, cmd
}

// save writes the note when its content changed since it was loaded and
// emits DoneMsg. On a write failure the editor stays open with the error.
func (m Model) save() (tea.Model, tea.Cmd) {
	now := notes.Normalize(m.ta.Value())
	if now != m.baseline {
		if err := m.store.Write(m.note.Name, now); err != nil {
			m.err = err.Error()
			return m, nil
		}
		m.baseline = now
	}
	return m, func() tea.Msg { return DoneMsg{} }
}

// statusBar renders the bottom line: hints left, note name right.
func (m Model) statusBar() string {
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
	body := m.ta.View()
	if m.err != "" {
		lines := strings.Split(body, "\n")
		lines[len(lines)-1] = theme.Error.Render(m.err)
		body = strings.Join(lines, "\n")
	}
	return body + "\n" + m.statusBar()
}
