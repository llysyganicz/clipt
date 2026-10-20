// Package editor implements the edit mode: the current note's raw markdown
// in a text area with keyboard selection but no modal editing. Esc saves the
// note when it changed and hands control back to the caller.
package editor

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/llysyganicz/clipt/internal/frontmatter"
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
	copyHint      = " · ctrl+shift+c copy" // appended while a selection exists
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
	// The text area's default keys, selection and copy included; the one
	// override is ctrl+t, which means templates everywhere, so the text
	// area must never transpose.
	ta.KeyMap = textarea.DefaultKeyMap()
	ta.KeyMap.TransposeCharacterBackward = key.NewBinding(key.WithDisabled())
	ta.SetStyles(theme.EditorStyles())
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
// grows while scrolled down. Reloading the buffer resets the offset to the
// top; the new size is applied before the cursor is put back where it was,
// because every cursor move scrolls the view just enough to keep the cursor
// visible at the current height.
func (m *Model) reflow(width, height int) {
	row := m.ta.Line()
	info := m.ta.LineInfo()
	col := info.StartColumn + info.ColumnOffset
	m.ta.SetValue(m.ta.Value())
	m.resize(width, height)
	for m.ta.Line() > row {
		m.ta.CursorUp()
	}
	m.ta.SetCursorColumn(col)
	m.scrollToCursor()
}

// Init returns the command that starts the cursor blinking.
func (m Model) Init() tea.Cmd { return textarea.Blink }

// Update handles a message and returns the new model.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.reflow(msg.Width, msg.Height)
		return m, nil
	case tea.KeyPressMsg:
		m.err = ""
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+q":
			return m.SaveAndQuit()
		case "esc":
			return m.done()
		case "ctrl+t":
			if m.template {
				return m, nil
			}
			return m, func() tea.Msg { return InsertTemplateMsg{} }
		case "tab":
			return m.Insert(indent), nil
		case "shift+tab":
			m.dedent()
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.ta, cmd = m.ta.Update(msg)
	// The text area reports clipboard failures in its Err field rather than
	// as a message; show them where write failures are shown.
	if m.ta.Err != nil {
		m.err = m.ta.Err.Error()
		m.ta.Err = nil
	}
	if m.syncGutterWidth() {
		m.scrollToCursor()
	}
	return m, cmd
}

// Insert inserts text at the cursor as if typed: a selection is replaced,
// the cursor ends after the text and the view scrolls to it.
func (m Model) Insert(text string) Model {
	m.ta.DeleteSelection()
	m.ta.InsertString(text)
	m.syncGutterWidth()
	m.scrollToCursor()
	return m
}

// InsertTemplate inserts a template's body at the cursor as Insert does
// (replacing a selection), then merges the template's tags into the note's front matter. The cursor
// stays after the inserted body, moved by the lines the merge added or
// removed above it. When the merge is refused the body stays inserted, the
// front matter is left as it was and the error is shown.
func (m Model) InsertTemplate(body string, tags []string) Model {
	m = m.Insert(body)
	if len(tags) == 0 {
		return m
	}
	merged, delta, err := frontmatter.MergeTags(m.ta.Value(), tags)
	if err != nil {
		msg, _, _ := strings.Cut(err.Error(), "\n")
		m.err = "tags not merged: " + msg
		return m
	}
	if merged == m.ta.Value() {
		return m
	}
	row := m.ta.Line() + delta
	info := m.ta.LineInfo()
	col := info.StartColumn + info.ColumnOffset
	m.ta.SetValue(merged)
	for m.ta.Line() > row {
		m.ta.CursorUp()
	}
	m.ta.SetCursorColumn(col)
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
	m.ta.SetCursorColumn(max(0, col-n))
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
func (m Model) done() (Model, tea.Cmd) {
	if !m.save() {
		return m, nil
	}
	return m, func() tea.Msg { return DoneMsg{Template: m.template} }
}

// SaveAndQuit saves and quits clipt; on a write failure the editor stays
// open with the error and clipt keeps running. The root model calls it when
// the user quits from the template insert mode over this editor.
func (m Model) SaveAndQuit() (Model, tea.Cmd) {
	if !m.save() {
		return m, nil
	}
	return m, tea.Quit
}

// statusBar renders the bottom line: hints left (with the copy key while a
// selection exists), the note name (or "template: <name>") right.
func (m Model) statusBar() string {
	keys, label := hints, m.note.Name
	if m.template {
		keys, label = templateHints, "template: "+m.note.Name
	}
	if m.ta.HasSelection() {
		keys += copyHint
	}
	left := theme.Hint.Render(keys)
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

// View renders the editor.
func (m Model) View() string {
	labels := m.gutterLabels()
	blank := strings.Repeat(" ", m.gutterW)
	m.ta.SetPromptFunc(m.gutterW, func(info textarea.PromptInfo) string {
		if info.LineNumber < len(labels) {
			return labels[info.LineNumber]
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
