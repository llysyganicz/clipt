// Package finder implements the fuzzy finder mode.
package finder

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/llysyganicz/clipt/internal/frontmatter"
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
// edit. Note.Tags holds the tags it was created with.
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
	width   int
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
	ti.SetStyles(theme.InputStyles())
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
	m.rows = BuildRows(m.all, m.current, ParseQuery(m.input.Value()))
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

// Init returns the command that starts the cursor blinking.
func (m Model) Init() tea.Cmd { return textinput.Blink }

// Update handles a message and returns the new model.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.width, m.height = ws.Width, ws.Height
		return m, nil
	}
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	m.err = ""
	switch key.String() {
	case "ctrl+c", "ctrl+q":
		return m, tea.Quit
	case "esc":
		return m, func() tea.Msg { return BackMsg{} }
	case "ctrl+o":
		return m, func() tea.Msg { return ChangeDirMsg{} }
	case "ctrl+t":
		return m, func() tea.Msg { return ToggleScopeMsg{} }
	case "up", "ctrl+p":
		if m.sel > 0 {
			m.sel--
		}
		return m, nil
	case "down", "ctrl+n":
		if m.sel < len(m.rows)-1 {
			m.sel++
		}
		return m, nil
	case "enter":
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

func (m Model) enter() (Model, tea.Cmd) {
	if len(m.rows) == 0 {
		return m, nil
	}
	row := m.rows[m.sel]
	if !row.IsCreate() {
		n := *row.Note
		return m, func() tea.Msg { return OpenNoteMsg{Note: n, Scope: m.scope} }
	}
	if err := m.create(row.Create, row.CreateTags); err != nil {
		m.err = err.Error()
		return m, nil
	}
	n := notes.Note{Name: row.Create, ModTime: time.Now(), Tags: row.CreateTags}
	return m, func() tea.Msg { return NoteCreatedMsg{Note: n, Scope: m.scope} }
}

// create makes the file for the create row: a note, empty or holding only
// front matter with the tags, or a template holding the skeleton with the
// tags in a templates directory created on demand. The name is validated
// before anything touches the disk.
func (m Model) create(name string, tags []string) error {
	if m.scope == ScopeNotes {
		if err := m.store.Create(name); err != nil || len(tags) == 0 {
			return err
		}
		return m.store.Write(name, frontmatter.Fence+"\n"+frontmatter.Line(tags)+"\n"+frontmatter.Fence+"\n")
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
	return m.store.Write(name, templates.SkeletonWithTags(tags))
}

// View renders the finder.
func (m Model) View() string {
	var b strings.Builder
	b.WriteString(m.input.View())
	if len(m.all) == 0 && m.input.Value() == "" {
		b.WriteString("\n" + theme.Hint.Render(m.emptyHint()))
	}
	start, end := m.window()
	for i := start; i < end; i++ {
		row := m.rows[i]
		marker, text, tags, style := "  ", "", []string(nil), theme.Normal
		if row.IsCreate() {
			text, tags, style = fmt.Sprintf("Create %q", row.Create), row.CreateTags, theme.Create
		} else {
			text, tags = row.Note.Name, row.Note.Tags
		}
		if i == m.sel {
			marker = "> "
			if !row.IsCreate() {
				style = theme.Selected
			}
		}
		b.WriteString("\n" + RenderRow(marker+text, style, tags, m.width))
	}
	if m.err != "" {
		b.WriteString("\n" + theme.Error.Render(m.err))
	}
	b.WriteString("\n" + theme.Hint.Render(m.footer()))
	return b.String()
}

// RenderRow renders a list row: head in style, then the tags dimmed in the
// tag display form. When width is known and the row does not fit, the tags
// are cut from the end and marked with "…", or left out when there is no
// room for them; head is never shortened.
func RenderRow(head string, style lipgloss.Style, tags []string, width int) string {
	out := style.Render(head)
	if len(tags) == 0 {
		return out
	}
	disp := frontmatter.Display(tags)
	if width > 0 {
		avail := width - lipgloss.Width(head) - 1
		switch {
		case lipgloss.Width(disp) <= avail:
		case avail >= 2:
			disp = ansi.Truncate(disp, avail, "…")
		default:
			return out
		}
	}
	return out + " " + theme.RowTag.Render(disp)
}

func (m Model) emptyHint() string {
	if m.scope == ScopeTemplates {
		return "No templates yet. Type a name and press Enter to create the first one."
	}
	return "No notes yet. Type a name and press Enter to create the first one."
}

func (m Model) footer() string {
	if m.scope == ScopeTemplates {
		return "enter edit · esc notes · ctrl+t notes · ctrl+o change dir · ctrl+q quit"
	}
	return "enter open · esc back · ctrl+t templates · ctrl+o change dir · ctrl+q quit"
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
