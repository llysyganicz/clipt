// Package finder implements the fuzzy finder mode.
package finder

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/llysyganicz/clipt/internal/notes"
	"github.com/llysyganicz/clipt/internal/templates"
	"github.com/llysyganicz/clipt/internal/ui/theme"
)

// Scope says what the finder lists: notes or templates.
type Scope int

const (
	// ScopeNotes lists the notes directory; the default.
	ScopeNotes Scope = iota
	// ScopeTemplates lists the templates under the notes directory.
	ScopeTemplates
)

// OpenNoteMsg asks the root to open Note: in preview for a note, in edit for
// a template.
type OpenNoteMsg struct {
	Note  notes.Note
	Scope Scope
}

// NoteCreatedMsg reports a freshly created note or template, to be opened in
// edit.
type NoteCreatedMsg struct {
	Note  notes.Note
	Scope Scope
}

// ChangeDirMsg asks the root to open the notes directory picker.
type ChangeDirMsg struct{}

// BackMsg asks the root to leave the finder: preview of current or quit in
// notes scope, back to notes scope in template scope.
type BackMsg struct{}

// ToggleScopeMsg asks the root to rebuild the finder in the other scope.
type ToggleScopeMsg struct{}

// Model is the finder Bubble Tea model.
type Model struct {
	store   notes.Store
	scope   Scope
	current *notes.Note
	all     []notes.Note
	rows    []Row
	input   textinput.Model
	sel     int
	height  int
	err     string
}

// New creates a finder over the notes in store. The listing is read from
// disk now.
func New(store notes.Store, current *notes.Note, query string) Model {
	return newModel(store, ScopeNotes, current, query)
}

// NewTemplates creates a finder in template scope over the templates store
// (see templates.Store). Nothing is hidden: templates never become the
// current note.
func NewTemplates(store notes.Store, query string) Model {
	return newModel(store, ScopeTemplates, nil, query)
}

func newModel(store notes.Store, scope Scope, current *notes.Note, query string) Model {
	ti := textinput.New()
	ti.Prompt = "> "
	if scope == ScopeTemplates {
		ti.Prompt = "templates> "
	}
	ti.PromptStyle = theme.Prompt
	ti.SetValue(query)
	ti.CursorEnd()
	ti.Focus()
	m := Model{store: store, scope: scope, current: current, input: ti}
	m.reload()
	return m
}

// Scope returns the finder's scope.
func (m Model) Scope() Scope { return m.scope }

func (m *Model) reload() {
	all, err := m.store.List()
	if err != nil && !(m.scope == ScopeTemplates && errors.Is(err, fs.ErrNotExist)) {
		m.err = err.Error()
	}
	m.all = all
	m.rebuild()
}

func (m *Model) rebuild() {
	m.rows = BuildRows(m.all, m.current, m.input.Value())
	m.sel = 0
}

// WithCurrent returns the finder with a new current note and an optional
// error, re-reading the listing.
func (m Model) WithCurrent(current *notes.Note, errMsg string) Model {
	m.current = current
	m.err = errMsg
	m.reload()
	return m
}

// Query returns the current query text.
func (m Model) Query() string { return m.input.Value() }

// Rows returns the visible rows.
func (m Model) Rows() []Row { return m.rows }

// Selected returns the selected row index.
func (m Model) Selected() int { return m.sel }

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return textinput.Blink }

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.height = ws.Height
		return m, nil
	}
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	m.err = ""
	switch key.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEsc:
		return m, func() tea.Msg { return BackMsg{} }
	case tea.KeyCtrlO:
		return m, func() tea.Msg { return ChangeDirMsg{} }
	case tea.KeyCtrlT:
		return m, func() tea.Msg { return ToggleScopeMsg{} }
	case tea.KeyUp, tea.KeyCtrlP:
		if m.sel > 0 {
			m.sel--
		}
		return m, nil
	case tea.KeyDown, tea.KeyCtrlN:
		if m.sel < len(m.rows)-1 {
			m.sel++
		}
		return m, nil
	case tea.KeyEnter:
		return m.enter()
	}
	before := m.input.Value()
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	if m.input.Value() != before {
		m.rebuild()
	}
	return m, cmd
}

func (m Model) enter() (tea.Model, tea.Cmd) {
	if len(m.rows) == 0 {
		return m, nil
	}
	row := m.rows[m.sel]
	if !row.IsCreate() {
		n := *row.Note
		return m, func() tea.Msg { return OpenNoteMsg{Note: n, Scope: m.scope} }
	}
	if err := m.create(row.Create); err != nil {
		m.err = err.Error()
		return m, nil
	}
	n := notes.Note{Name: row.Create, ModTime: time.Now()}
	return m, func() tea.Msg { return NoteCreatedMsg{Note: n, Scope: m.scope} }
}

// create makes the file for the create row: an empty note, or a template
// holding the skeleton in a templates directory created on demand. The name
// is validated before anything touches the disk.
func (m Model) create(name string) error {
	if m.scope == ScopeNotes {
		return m.store.Create(name)
	}
	if err := notes.ValidateName(name); err != nil {
		return err
	}
	if err := templates.EnsureDir(filepath.Dir(m.store.Dir)); err != nil {
		return err
	}
	if err := m.store.Create(name); err != nil {
		return err
	}
	return m.store.Write(name, templates.Skeleton)
}

// View implements tea.Model.
func (m Model) View() string {
	var b strings.Builder
	b.WriteString(m.input.View())
	if len(m.all) == 0 && m.input.Value() == "" {
		b.WriteString("\n" + theme.Hint.Render(m.emptyHint()))
	}
	start, end := m.window()
	for i := start; i < end; i++ {
		row := m.rows[i]
		marker, text, style := "  ", "", theme.Normal
		if row.IsCreate() {
			text, style = fmt.Sprintf("Create %q", row.Create), theme.Create
		} else {
			text = row.Note.Name
		}
		if i == m.sel {
			marker = "> "
			if !row.IsCreate() {
				style = theme.Selected
			}
		}
		b.WriteString("\n" + style.Render(marker+text))
	}
	if m.err != "" {
		b.WriteString("\n" + theme.Error.Render(m.err))
	}
	b.WriteString("\n" + theme.Hint.Render(m.footer()))
	return b.String()
}

func (m Model) emptyHint() string {
	if m.scope == ScopeTemplates {
		return "No templates yet. Type a name and press Enter to create the first one."
	}
	return "No notes yet. Type a name and press Enter to create the first one."
}

func (m Model) footer() string {
	if m.scope == ScopeTemplates {
		return "enter edit · esc notes · ctrl+t notes · ctrl+o change directory"
	}
	return "enter open · esc back · ctrl+t templates · ctrl+o change directory"
}

// window returns the slice of rows that fits the terminal height.
func (m Model) window() (int, int) {
	n := len(m.rows)
	avail := m.height - 4 // query, hint/empty line, error line, footer
	if m.height == 0 || avail <= 0 || n <= avail {
		return 0, n
	}
	start := m.sel - avail + 1
	if start < 0 {
		start = 0
	}
	return start, start + avail
}
