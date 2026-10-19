// Package app holds the root Bubble Tea model: it routes between modes and
// owns the optional "current note".
package app

import (
	"fmt"
	"path/filepath"

	tea "charm.land/bubbletea/v2"

	"github.com/llysyganicz/clipt/internal/notes"
	"github.com/llysyganicz/clipt/internal/templates"
	"github.com/llysyganicz/clipt/internal/ui/editor"
	"github.com/llysyganicz/clipt/internal/ui/finder"
	"github.com/llysyganicz/clipt/internal/ui/picker"
	"github.com/llysyganicz/clipt/internal/ui/preview"
	"github.com/llysyganicz/clipt/internal/ui/tmpl"
)

type mode int

const (
	modePicker mode = iota
	modeFinder
	modePreview
	modeEdit
	modeTemplate // template insert: picker and form over the kept editor
)

// Model is the root model.
type Model struct {
	mode       mode
	dir        string
	current    *notes.Note
	pickerFrom picker.Origin

	picker  picker.Model
	finder  finder.Model
	preview preview.Model
	edit    editor.Model
	tmpl    tmpl.Model
	width   int
	height  int
}

// New starts in the finder over dir.
func New(dir string) Model {
	m := Model{mode: modeFinder, dir: dir}
	m.finder = finder.New(notes.Store{Dir: dir}, nil, "")
	return m
}

// NewPicker starts in the picker because no usable directory is configured.
// prefill and errMsg describe a stale configured path, if any.
func NewPicker(prefill, errMsg string) Model {
	if prefill == "" && errMsg == "" {
		prefill = picker.DefaultPath
	}
	return Model{
		mode:       modePicker,
		pickerFrom: picker.Startup,
		picker:     picker.New(picker.Startup, prefill, errMsg),
	}
}

// Dir returns the active notes directory.
func (m Model) Dir() string { return m.dir }

// Current returns the current note, or nil.
func (m Model) Current() *notes.Note { return m.current }

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	if m.mode == modePicker {
		return m.picker.Init()
	}
	return m.finder.Init()
}

func (m Model) store() notes.Store { return notes.Store{Dir: m.dir} }

func (m Model) templatesStore() notes.Store { return templates.Store(m.dir) }

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.finder, _ = m.finder.Update(msg)
		switch m.mode {
		case modePreview:
			m.preview, _ = m.preview.Update(msg)
		case modeEdit:
			m.edit, _ = m.edit.Update(msg)
		case modeTemplate:
			// The editor is kept behind the insert mode and must be the
			// right size when it reappears.
			m.edit, _ = m.edit.Update(msg)
			m.tmpl, _ = m.tmpl.Update(msg)
		}
		return m, nil

	case finder.OpenNoteMsg:
		if msg.Scope == finder.ScopeTemplates {
			return m.openTemplate(msg.Note)
		}
		return m.showPreview(msg.Note)
	case finder.NoteCreatedMsg:
		if msg.Scope == finder.ScopeTemplates {
			return m.openTemplate(msg.Note)
		}
		out, cmd := m.openEdit(msg.Note)
		m = out.(Model)
		if m.mode == modeEdit && len(msg.Note.Tags) > 0 {
			// The note is only front matter; start on the line after it.
			// Saving trims the newline back, so an untouched note is kept.
			m.edit = m.edit.Insert("\n")
		}
		return m, cmd
	case finder.ToggleScopeMsg:
		if m.finder.Scope() == finder.ScopeTemplates {
			return m.toFinder()
		}
		return m.toTemplatesFinder()

	case preview.BackMsg:
		return m.toFinder()
	case preview.EditMsg:
		return m.openEdit(msg.Note)
	case preview.RenamedMsg:
		n := msg.Note
		m.current = &n
		return m, nil
	case preview.DeletedMsg:
		m.current = nil
		return m.toFinder()

	case finder.ChangeDirMsg:
		m.pickerFrom = picker.Finder
		m.picker = picker.New(picker.Finder, m.dir, "")
		m.mode = modePicker
		return m, m.picker.Init()

	case finder.BackMsg:
		if m.finder.Scope() == finder.ScopeTemplates {
			return m.toFinder()
		}
		return m.back()

	case picker.DirAcceptedMsg:
		return m.dirAccepted(msg.Dir)

	case picker.PickerCancelledMsg:
		if m.pickerFrom == picker.Startup {
			return m, tea.Quit
		}
		if m.finder.Scope() == finder.ScopeTemplates {
			return m.toFinder()
		}
		m.mode = modeFinder
		return m, nil

	case editor.DoneMsg:
		if msg.Template {
			return m.toTemplatesFinder()
		}
		if m.current == nil {
			return m.toFinder()
		}
		return m.showPreview(*m.current)
	case editor.InsertTemplateMsg:
		m.tmpl = tmpl.New(m.templatesStore())
		if m.width > 0 && m.height > 0 {
			m.tmpl, _ = m.tmpl.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		}
		m.mode = modeTemplate
		return m, m.tmpl.Init()

	case tmpl.InsertMsg:
		m.edit = m.edit.InsertTemplate(msg.Text, msg.Tags)
		m.mode = modeEdit
		return m, nil
	case tmpl.CancelMsg:
		m.mode = modeEdit
		return m, nil
	case tmpl.QuitMsg:
		// Abandon the insertion; the note's own edits are saved by the
		// editor, which stays open with the error when that fails.
		m.mode = modeEdit
		var cmd tea.Cmd
		m.edit, cmd = m.edit.SaveAndQuit()
		return m, cmd
	}

	var cmd tea.Cmd
	switch m.mode {
	case modePicker:
		m.picker, cmd = m.picker.Update(msg)
	case modeFinder:
		m.finder, cmd = m.finder.Update(msg)
	case modePreview:
		m.preview, cmd = m.preview.Update(msg)
	case modeEdit:
		m.edit, cmd = m.edit.Update(msg)
	case modeTemplate:
		m.tmpl, cmd = m.tmpl.Update(msg)
	}
	return m, cmd
}

func (m Model) newFinder(query string) finder.Model {
	return m.sized(finder.New(m.store(), m.current, query))
}

func (m Model) newTemplatesFinder() finder.Model {
	return m.sized(finder.NewTemplates(m.templatesStore(), ""))
}

func (m Model) sized(f finder.Model) finder.Model {
	if m.height > 0 {
		f, _ = f.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
	}
	return f
}

// toFinder enters the finder in notes scope with an empty query.
func (m Model) toFinder() (tea.Model, tea.Cmd) {
	m.finder = m.newFinder("")
	m.mode = modeFinder
	return m, m.finder.Init()
}

// toTemplatesFinder enters the finder in template scope with an empty query.
func (m Model) toTemplatesFinder() (tea.Model, tea.Cmd) {
	m.finder = m.newTemplatesFinder()
	m.mode = modeFinder
	return m, m.finder.Init()
}

// showPreview makes n the current note and opens its preview. When the note
// cannot be read the finder stays open with an error and no current note.
func (m Model) showPreview(n notes.Note) (tea.Model, tea.Cmd) {
	pv, err := preview.New(m.store(), n)
	if err != nil {
		m.current = nil
		m.finder = m.finder.WithCurrent(nil, fmt.Sprintf("cannot read %q", n.Name))
		m.mode = modeFinder
		return m, nil
	}
	if m.width > 0 && m.height > 0 {
		pv, _ = pv.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
	}
	m.current = &n
	m.preview = pv
	m.mode = modePreview
	return m, nil
}

// openEdit makes n the current note and opens the editor on it. When the note
// cannot be read the finder stays open with an error and no current note.
func (m Model) openEdit(n notes.Note) (tea.Model, tea.Cmd) {
	ed, err := editor.New(m.store(), n)
	if err != nil {
		m.current = nil
		m.finder = m.finder.WithCurrent(nil, fmt.Sprintf("cannot read %q", n.Name))
		m.mode = modeFinder
		return m, nil
	}
	if m.width > 0 && m.height > 0 {
		ed, _ = ed.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
	}
	m.current = &n
	m.edit = ed
	m.mode = modeEdit
	return m, ed.Init()
}

// openTemplate opens the editor on a template. The current note is not
// touched; when the template cannot be read the finder stays in template
// scope with an error.
func (m Model) openTemplate(n notes.Note) (tea.Model, tea.Cmd) {
	ed, err := editor.NewTemplate(m.templatesStore(), n)
	if err != nil {
		m.finder = m.finder.WithCurrent(nil, fmt.Sprintf("cannot read %q", n.Name))
		m.mode = modeFinder
		return m, nil
	}
	if m.width > 0 && m.height > 0 {
		ed, _ = ed.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
	}
	m.edit = ed
	m.mode = modeEdit
	return m, ed.Init()
}

func (m Model) back() (tea.Model, tea.Cmd) {
	if m.current == nil {
		return m, tea.Quit
	}
	if !m.store().Exists(m.current.Name) {
		name := m.current.Name
		m.current = nil
		m.finder = m.finder.WithCurrent(nil, fmt.Sprintf("note %q no longer exists", name))
		return m, nil
	}
	return m.showPreview(*m.current)
}

func (m Model) dirAccepted(dir string) (tea.Model, tea.Cmd) {
	if m.pickerFrom == picker.Finder && filepath.Clean(dir) != filepath.Clean(m.dir) {
		m.current = nil
	}
	m.dir = dir
	m.finder = m.newFinder("")
	m.mode = modeFinder
	return m, m.finder.Init()
}

// View implements tea.Model: the active mode's rendering, in the alternate
// screen.
func (m Model) View() tea.View {
	v := tea.NewView(m.content())
	v.AltScreen = true
	return v
}

func (m Model) content() string {
	switch m.mode {
	case modePicker:
		return m.picker.View()
	case modePreview:
		return m.preview.View()
	case modeEdit:
		return m.edit.View()
	case modeTemplate:
		return m.tmpl.View()
	}
	return m.finder.View()
}
