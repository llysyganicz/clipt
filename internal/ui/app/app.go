// Package app holds the root Bubble Tea model: it routes between modes and
// owns the optional "current note".
package app

import (
	"fmt"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/llysyganicz/clipt/internal/notes"
	"github.com/llysyganicz/clipt/internal/ui/finder"
	"github.com/llysyganicz/clipt/internal/ui/picker"
)

type mode int

const (
	modePicker mode = iota
	modeFinder
	modeNote // placeholder for preview/edit
)

// Model is the root model.
type Model struct {
	mode       mode
	dir        string
	current    *notes.Note
	pickerFrom picker.Origin

	picker picker.Model
	finder finder.Model
	note   placeholder
	height int
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

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		nf, _ := m.finder.Update(msg)
		m.finder = nf.(finder.Model)
		return m, nil

	case finder.OpenNoteMsg:
		return m.openNote(msg.Note), nil
	case finder.NoteCreatedMsg:
		return m.openNote(msg.Note), nil

	case finder.ChangeDirMsg:
		m.pickerFrom = picker.Finder
		m.picker = picker.New(picker.Finder, m.dir, "")
		m.mode = modePicker
		return m, m.picker.Init()

	case finder.BackMsg:
		return m.back()

	case picker.DirAcceptedMsg:
		return m.dirAccepted(msg.Dir)

	case picker.PickerCancelledMsg:
		if m.pickerFrom == picker.Startup {
			return m, tea.Quit
		}
		m.mode = modeFinder
		return m, nil

	case noteBackMsg:
		m.finder = m.newFinder("")
		m.mode = modeFinder
		return m, m.finder.Init()
	}

	var cmd tea.Cmd
	switch m.mode {
	case modePicker:
		var nm tea.Model
		nm, cmd = m.picker.Update(msg)
		m.picker = nm.(picker.Model)
	case modeFinder:
		var nm tea.Model
		nm, cmd = m.finder.Update(msg)
		m.finder = nm.(finder.Model)
	case modeNote:
		var nm tea.Model
		nm, cmd = m.note.Update(msg)
		m.note = nm.(placeholder)
	}
	return m, cmd
}

func (m Model) newFinder(query string) finder.Model {
	f := finder.New(m.store(), m.current, query)
	if m.height > 0 {
		nf, _ := f.Update(tea.WindowSizeMsg{Height: m.height})
		f = nf.(finder.Model)
	}
	return f
}

func (m Model) openNote(n notes.Note) Model {
	m.current = &n
	m.note = placeholder{name: n.Name}
	m.mode = modeNote
	return m
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
	m.note = placeholder{name: m.current.Name}
	m.mode = modeNote
	return m, nil
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

// View implements tea.Model.
func (m Model) View() string {
	switch m.mode {
	case modePicker:
		return m.picker.View()
	case modeNote:
		return m.note.View()
	}
	return m.finder.View()
}
