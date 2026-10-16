// Package editor implements the edit mode: the current note's raw markdown
// in an insert-only text area. Esc saves the note when it changed and hands
// control back to the caller.
package editor

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/llysyganicz/clipt/internal/notes"
	"github.com/llysyganicz/clipt/internal/ui/theme"
)

// DoneMsg is emitted on Esc once the file is saved (or was unchanged).
// Template says whether a template rather than a note was being edited.
type DoneMsg struct{ Template bool }

// InsertTemplateMsg is emitted on ctrl+t while editing a note: the root
// should open the template picker and hand the result to Insert.
type InsertTemplateMsg struct{}

const (
	defaultWidth  = 80
	defaultHeight = 24
	indent        = "  " // what Tab inserts and Shift+Tab removes
	hints         = "esc save & back · ctrl+t template · ctrl+q save & quit"
	templateHints = "esc save & back · ctrl+q save & quit"
)

// repositionMsg is a no-op message sent through the text area so it scrolls
// its viewport to the cursor.
type repositionMsg struct{}

// Model is the editor Bubble Tea model.
type Model struct {
	store    notes.Store
	note     notes.Note
	template bool
	ta       textarea.Model
	gutterW  int // columns reserved for the line number gutter
	baseline string
	width    int
	height   int
	err      string
}

// New reads the note and returns an editor on it with the cursor at the end.
// It fails when the note cannot be read so the caller can fall back to the
// finder.
func New(store notes.Store, note notes.Note) (Model, error) {
	return load(store, note, false)
}

// NewTemplate is New for a template: store is the templates store, ctrl+t
// does nothing, and DoneMsg reports Template.
func NewTemplate(store notes.Store, note notes.Note) (Model, error) {
	return load(store, note, true)
}

func load(store notes.Store, note notes.Note, template bool) (Model, error) {
	content, err := store.Read(note.Name)
	if err != nil {
		return Model{}, err
	}
	content = strings.ReplaceAll(content, "\r\n", "\n")
	// The on-disk form ends with one newline; without dropping it the text
	// area would show an extra empty last line and put the cursor there.
	content = strings.TrimSuffix(content, "\n")

	ta := textarea.New()
	ta.ShowLineNumbers = false // the editor draws its own gutter, see gutter.go
	ta.Prompt = ""
	ta.CharLimit = 0
	ta.MaxHeight = 0
	ta.MaxWidth = 0
	ta.KeyMap = textarea.DefaultKeyMap
	// ctrl+t means templates everywhere; never let the text area transpose.
	ta.KeyMap.TransposeCharacterBackward = key.NewBinding(key.WithDisabled())
	ta.FocusedStyle, ta.BlurredStyle = theme.EditorStyles()
	ta.Focus()
	ta.SetValue(content)

	m := Model{
		store:    store,
		note:     note,
		template: template,
		ta:       ta,
		baseline: notes.Normalize(ta.Value()),
	}
	m.syncGutterWidth()
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
		case tea.KeyCtrlQ:
			return m.SaveAndQuit()
		case tea.KeyEsc:
			return m.done()
		case tea.KeyCtrlT:
			if m.template {
				return m, nil
			}
			return m, func() tea.Msg { return InsertTemplateMsg{} }
		case tea.KeyTab:
			return m.Insert(indent), nil
		case tea.KeyShiftTab:
			m.dedent()
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.ta, cmd = m.ta.Update(msg)
	if m.syncGutterWidth() {
		m.scrollToCursor()
	}
	return m, cmd
}

// Insert inserts text at the cursor as if typed, leaves the cursor after it
// and scrolls the view to it.
func (m Model) Insert(text string) Model {
	m.ta.InsertString(text)
	m.syncGutterWidth()
	m.scrollToCursor()
	return m
}

// dedent removes up to len(indent) spaces from the start of the cursor's
// line and moves the cursor left by as many, not before the line start. The
// text area has no line-level setter, so the whole value is replaced and
// the cursor put back, as reflow does.
func (m *Model) dedent() {
	row := m.ta.Line()
	info := m.ta.LineInfo()
	col := info.StartColumn + info.ColumnOffset
	lines := strings.Split(m.ta.Value(), "\n")
	n := 0
	for n < len(indent) && n < len(lines[row]) && lines[row][n] == ' ' {
		n++
	}
	if n == 0 {
		return
	}
	lines[row] = lines[row][n:]
	m.ta.SetValue(strings.Join(lines, "\n"))
	for m.ta.Line() > row {
		m.ta.CursorUp()
	}
	m.ta.SetCursor(max(0, col-n))
	m.scrollToCursor()
}

// save writes the note when its content changed since it was loaded and
// reports whether the editor may be left. On a write failure it records the
// error and reports false.
func (m *Model) save() bool {
	now := notes.Normalize(m.ta.Value())
	if now == m.baseline {
		return true
	}
	if err := m.store.Write(m.note.Name, now); err != nil {
		m.err = err.Error()
		return false
	}
	m.baseline = now
	return true
}

// done saves and emits DoneMsg; on a write failure the editor stays open
// with the error.
func (m Model) done() (tea.Model, tea.Cmd) {
	if !m.save() {
		return m, nil
	}
	return m, func() tea.Msg { return DoneMsg{Template: m.template} }
}

// SaveAndQuit saves and quits clipt; on a write failure the editor stays
// open with the error and clipt keeps running. The root model calls it when
// the user quits from the template insert mode over this editor.
func (m Model) SaveAndQuit() (tea.Model, tea.Cmd) {
	if !m.save() {
		return m, nil
	}
	return m, tea.Quit
}

// statusBar renders the bottom line: hints left, the note name (or
// "template: <name>") right.
func (m Model) statusBar() string {
	left := theme.Hint.Render(hints)
	label := m.note.Name
	if m.template {
		left = theme.Hint.Render(templateHints)
		label = "template: " + m.note.Name
	}
	avail := m.width - lipgloss.Width(left) - 1
	name := []rune(label)
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
	labels := m.gutterLabels()
	blank := strings.Repeat(" ", m.gutterW)
	m.ta.SetPromptFunc(m.gutterW, func(row int) string {
		if row < len(labels) {
			return labels[row]
		}
		return blank
	})
	body := m.ta.View()
	if m.err != "" {
		lines := strings.Split(body, "\n")
		lines[len(lines)-1] = theme.Error.Render(m.err)
		body = strings.Join(lines, "\n")
	}
	return body + "\n" + m.statusBar()
}
