// Package picker implements the notes directory selection mode.
package picker

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/llysyganicz/clipt/internal/config"
	"github.com/llysyganicz/clipt/internal/ui/theme"
)

// Origin records which entry point opened the picker.
type Origin int

const (
	// Startup means the picker was opened because no directory is configured.
	Startup Origin = iota
	// Finder means the picker was opened with ctrl+o from the finder.
	Finder
)

// DefaultPath is the prefill on first run.
const DefaultPath = "~/notes"

// DirAcceptedMsg is emitted once a directory is chosen and saved to the config.
type DirAcceptedMsg struct{ Dir string }

// PickerCancelledMsg is emitted when the user presses Esc in the path input.
type PickerCancelledMsg struct{}

type state int

const (
	stateInput state = iota
	stateConfirm
)

// Model is the picker Bubble Tea model.
type Model struct {
	origin  Origin
	state   state
	input   textinput.Model
	pending string // absolute path awaiting create confirmation
	err     string
}

// New creates a picker. prefill is the initial input text and errMsg an
// optional error shown until the first keypress.
func New(origin Origin, prefill, errMsg string) Model {
	ti := textinput.New()
	ti.Prompt = "Notes directory: "
	ti.PromptStyle = theme.Prompt
	ti.SetValue(prefill)
	ti.CursorEnd()
	ti.Focus()
	return Model{origin: origin, input: ti, err: errMsg}
}

// Origin returns the entry point that opened the picker.
func (m Model) Origin() Origin { return m.origin }

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return textinput.Blink }

// expandPath trims, expands a leading ~ using home, and makes the path absolute.
func expandPath(in, home string) (string, error) {
	p := strings.TrimSpace(in)
	if p == "" {
		return "", errors.New("a path is required")
	}
	if p == "~" {
		p = home
	} else if strings.HasPrefix(p, "~/") {
		p = filepath.Join(home, p[2:])
	}
	return filepath.Abs(p)
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	m.err = ""
	if key.Type == tea.KeyCtrlC || key.Type == tea.KeyCtrlQ {
		return m, tea.Quit
	}
	if m.state == stateConfirm {
		return m.updateConfirm(key)
	}
	switch key.Type {
	case tea.KeyEsc:
		return m, func() tea.Msg { return PickerCancelledMsg{} }
	case tea.KeyEnter:
		return m.submit()
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) submit() (tea.Model, tea.Cmd) {
	home, _ := os.UserHomeDir()
	dir, err := expandPath(m.input.Value(), home)
	if err != nil {
		m.err = err.Error()
		return m, nil
	}
	info, err := os.Stat(dir)
	switch {
	case err == nil && info.IsDir():
		return m.accept(dir)
	case err == nil:
		m.err = fmt.Sprintf("not a directory: %s", dir)
	case errors.Is(err, os.ErrNotExist):
		m.pending = dir
		m.state = stateConfirm
	default:
		m.err = err.Error()
	}
	return m, nil
}

func (m Model) updateConfirm(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Type == tea.KeyEsc, key.String() == "n", key.String() == "N":
		m.state = stateInput
	case key.String() == "y", key.String() == "Y":
		m.state = stateInput
		if err := os.MkdirAll(m.pending, 0o755); err != nil {
			m.err = err.Error()
			return m, nil
		}
		return m.accept(m.pending)
	}
	return m, nil
}

// accept saves the config; on failure the picker stays open with the error.
func (m Model) accept(dir string) (tea.Model, tea.Cmd) {
	if err := config.Save(dir); err != nil {
		m.state = stateInput
		m.err = err.Error()
		return m, nil
	}
	return m, func() tea.Msg { return DirAcceptedMsg{Dir: dir} }
}

// View implements tea.Model.
func (m Model) View() string {
	var b strings.Builder
	if m.state == stateConfirm {
		b.WriteString(theme.Prompt.Render(fmt.Sprintf("Create %s? (y/n)", m.pending)))
		return b.String()
	}
	b.WriteString(m.input.View())
	if m.err != "" {
		b.WriteString("\n" + theme.Error.Render(m.err))
	}
	b.WriteString("\n" + theme.Hint.Render("enter accept · esc back · ctrl+q quit"))
	return b.String()
}
