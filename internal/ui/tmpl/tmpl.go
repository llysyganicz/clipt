// Package tmpl implements the template insert mode entered from the editor:
// a fuzzy picker over the templates, then a form with one input per field.
// It knows nothing about the editor; it only reports the expanded text to
// insert, or that the user backed out.
package tmpl

import (
	"errors"
	"fmt"
	"io/fs"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/llysyganicz/clipt/internal/notes"
	"github.com/llysyganicz/clipt/internal/templates"
	"github.com/llysyganicz/clipt/internal/ui/finder"
	"github.com/llysyganicz/clipt/internal/ui/theme"
)

// InsertMsg hands the expanded template body and the template's tags back
// to the caller.
type InsertMsg struct {
	Text string
	Tags []string
}

// CancelMsg reports that the user left the insert mode without a result.
type CancelMsg struct{}

// QuitMsg is emitted on ctrl+q: the root should abandon the insertion, save
// the note through the editor and quit.
type QuitMsg struct{}

type state int

const (
	statePick state = iota
	stateForm
)

const (
	defaultWidth  = 80
	defaultHeight = 24

	pickHints  = "enter pick · esc back · ctrl+q quit"
	formHints  = "enter next/submit · tab/shift+tab move · ←/→ option · esc back · ctrl+q quit"
	emptyHint  = "No templates yet. Create one from the finder with ctrl+t."
	optionOpen = "< "
	optionEnd  = " >"
)

// row is one form input: a text or number input, or an option selector.
type row struct {
	field templates.Field
	input textinput.Model // text and number rows
	idx   int             // selected option for option rows
}

func (r row) value() string {
	if r.field.Type == templates.TypeOption {
		return r.field.Options[r.idx]
	}
	return r.input.Value()
}

// Model is the insert mode Bubble Tea model.
type Model struct {
	store notes.Store
	state state

	// picker
	query textinput.Model
	all   []notes.Note
	rows  []notes.Note
	sel   int

	// form
	tpl   templates.Template
	form  []row
	focus int

	width  int
	height int
	err    string
}

// New opens the picker over the templates in store (see templates.Store).
// The listing is read now; a missing directory lists as empty.
func New(store notes.Store) Model {
	ti := textinput.New()
	ti.Prompt = "template> "
	ti.SetStyles(theme.InputStyles())
	ti.Focus()
	m := Model{store: store, query: ti, width: defaultWidth, height: defaultHeight}
	all, err := store.List()
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		m.err = err.Error()
	}
	m.all = all
	m.filter()
	return m
}

func (m *Model) filter() {
	m.rows = m.rows[:0]
	q := finder.ParseQuery(m.query.Value())
	for _, n := range m.all {
		if finder.MatchItem(q, n) {
			m.rows = append(m.rows, n)
		}
	}
	m.sel = 0
}

// Rows returns the templates currently listed by the picker.
func (m Model) Rows() []notes.Note { return m.rows }

// Selected returns the selected picker row.
func (m Model) Selected() int { return m.sel }

// Query returns the picker's query text.
func (m Model) Query() string { return m.query.Value() }

// InForm reports whether the form rather than the picker is shown.
func (m Model) InForm() bool { return m.state == stateForm }

// Focus returns the focused form row.
func (m Model) Focus() int { return m.focus }

// Values returns the form's current values by field name.
func (m Model) Values() map[string]string {
	out := make(map[string]string, len(m.form))
	for _, r := range m.form {
		out[r.field.Name] = r.value()
	}
	return out
}

// Err returns the error shown, if any.
func (m Model) Err() string { return m.err }

// Init returns the command that starts the cursor blinking.
func (m Model) Init() tea.Cmd { return textinput.Blink }

// Update handles a message and returns the new model.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyPressMsg:
		m.err = ""
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+q":
			return m, func() tea.Msg { return QuitMsg{} }
		case "esc":
			return m, func() tea.Msg { return CancelMsg{} }
		}
		if m.state == stateForm {
			return m.updateForm(msg)
		}
		return m.updatePick(msg)
	}
	var cmd tea.Cmd
	if m.state == stateForm {
		m.form[m.focus].input, cmd = m.form[m.focus].input.Update(msg)
	} else {
		m.query, cmd = m.query.Update(msg)
	}
	return m, cmd
}

func (m Model) updatePick(key tea.KeyPressMsg) (Model, tea.Cmd) {
	switch key.String() {
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
		return m.pick()
	}
	before := m.query.Value()
	var cmd tea.Cmd
	m.query, cmd = m.query.Update(key)
	if m.query.Value() != before {
		m.filter()
	}
	return m, cmd
}

// pick reads and parses the selected template: a field-less one is inserted
// at once, otherwise the form opens. Errors are shown in place.
func (m Model) pick() (Model, tea.Cmd) {
	if len(m.rows) == 0 {
		return m, nil
	}
	name := m.rows[m.sel].Name
	content, err := m.store.Read(name)
	if err != nil {
		m.err = fmt.Sprintf("cannot read %q: %v", name, err)
		return m, nil
	}
	tpl, err := templates.Parse(content)
	if err != nil {
		m.err = fmt.Sprintf("%s: %v", name, err)
		return m, nil
	}
	if len(tpl.Fields) == 0 {
		return m, func() tea.Msg { return InsertMsg{Text: tpl.Body, Tags: tpl.Tags} }
	}
	m.tpl = tpl
	m.form = make([]row, len(tpl.Fields))
	for i, f := range tpl.Fields {
		// Option rows carry an input too so focus handling is uniform; they
		// never show or edit it.
		ti := textinput.New()
		ti.Prompt = ""
		ti.SetStyles(theme.InputStyles())
		ti.SetValue(f.Default())
		ti.CursorEnd()
		m.form[i] = row{field: f, input: ti}
	}
	m.focus = 0
	m.form[0].input.Focus()
	m.state = stateForm
	return m, textinput.Blink
}

func (m Model) updateForm(key tea.KeyPressMsg) (Model, tea.Cmd) {
	switch k := key.String(); k {
	case "enter":
		if m.focus == len(m.form)-1 {
			return m.submit()
		}
		return m.moveFocus(1)
	case "down", "tab":
		return m.moveFocus(1)
	case "up", "shift+tab":
		return m.moveFocus(-1)
	case "left", "right":
		r := &m.form[m.focus]
		if r.field.Type != templates.TypeOption {
			break
		}
		if k == "left" && r.idx > 0 {
			r.idx--
		}
		if k == "right" && r.idx < len(r.field.Options)-1 {
			r.idx++
		}
		return m, nil
	}
	if m.form[m.focus].field.Type == templates.TypeOption {
		return m, nil
	}
	var cmd tea.Cmd
	m.form[m.focus].input, cmd = m.form[m.focus].input.Update(key)
	return m, cmd
}

// moveFocus moves the focus by delta rows, stopping at the ends.
func (m Model) moveFocus(delta int) (Model, tea.Cmd) {
	next := m.focus + delta
	if next < 0 || next >= len(m.form) {
		return m, nil
	}
	return m.setFocus(next), textinput.Blink
}

func (m Model) setFocus(i int) Model {
	m.form[m.focus].input.Blur()
	m.focus = i
	m.form[i].input.Focus()
	return m
}

// submit validates the number rows and hands back the expanded body.
func (m Model) submit() (Model, tea.Cmd) {
	for i, r := range m.form {
		if r.field.Type != templates.TypeNumber {
			continue
		}
		if _, err := strconv.ParseFloat(strings.TrimSpace(r.value()), 64); err != nil {
			m.err = fmt.Sprintf("%s: %q is not a number", r.field.Name, r.value())
			return m.setFocus(i), nil
		}
	}
	text, tags := m.tpl.Expand(m.Values()), m.tpl.Tags
	return m, func() tea.Msg { return InsertMsg{Text: text, Tags: tags} }
}

// View renders the picker or the form.
func (m Model) View() string {
	var lines []string
	if m.state == stateForm {
		lines = m.formLines()
	} else {
		lines = m.pickLines()
	}
	if m.err != "" {
		lines = append(lines, theme.Error.Render(m.err))
	}
	hints := pickHints
	if m.state == stateForm {
		hints = formHints
	}
	for len(lines) < m.height-1 {
		lines = append(lines, "")
	}
	lines = append(lines, theme.Hint.Render(hints))
	return strings.Join(lines, "\n")
}

func (m Model) pickLines() []string {
	lines := []string{m.query.View()}
	if len(m.all) == 0 && m.query.Value() == "" {
		lines = append(lines, theme.Hint.Render(emptyHint))
	}
	start, end := m.window()
	for i := start; i < end; i++ {
		marker, style := "  ", theme.Normal
		if i == m.sel {
			marker, style = "> ", theme.Selected
		}
		lines = append(lines, finder.RenderRow(marker+m.rows[i].Name, style, m.rows[i].Tags, m.width))
	}
	return lines
}

// window returns the slice of picker rows that fits the terminal height.
func (m Model) window() (int, int) {
	n := len(m.rows)
	avail := m.height - 4 // query, hint/empty line, error line, footer
	if avail <= 0 || n <= avail {
		return 0, n
	}
	start := m.sel - avail + 1
	if start < 0 {
		start = 0
	}
	return start, start + avail
}

func (m Model) formLines() []string {
	lines := make([]string, 0, len(m.form))
	for i, r := range m.form {
		marker := "  "
		label := theme.Normal.Render(r.field.Label + ": ")
		if i == m.focus {
			marker = "> "
			label = theme.Prompt.Render(r.field.Label + ": ")
		}
		var value string
		if r.field.Type == templates.TypeOption {
			parts := make([]string, len(r.field.Options))
			for j, o := range r.field.Options {
				if j == r.idx {
					parts[j] = theme.Selected.Render(optionOpen + o + optionEnd)
				} else {
					parts[j] = theme.Hint.Render(o)
				}
			}
			value = strings.Join(parts, " ")
		} else {
			value = r.input.View()
		}
		lines = append(lines, marker+label+value)
	}
	return lines
}
