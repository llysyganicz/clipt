package tmpl

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/llysyganicz/clipt/internal/notes"
	"github.com/llysyganicz/clipt/internal/templates"
	"github.com/llysyganicz/clipt/internal/ui/uitest"
)

const withFields = "---\nform:\n  - name: topic\n    label: Topic\n    type: text\n  - name: count\n    type: number\n  - name: prio\n    type: option\n    options: [low, high]\n---\n# {{topic}} {{count}} {{prio}}\n"

// setup writes templates (later names are newer) under a notes dir's
// .templates and returns the templates store. With withDir false the
// directory is not created.
func setup(t *testing.T, files []string, contents map[string]string) notes.Store {
	t.Helper()
	store := templates.Store(t.TempDir())
	if files == nil && contents == nil {
		return store
	}
	if err := os.Mkdir(store.Dir, 0o755); err != nil {
		t.Fatal(err)
	}
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i, n := range files {
		c, ok := contents[n]
		if !ok {
			c = "# " + n + "\n"
		}
		p := filepath.Join(store.Dir, n+".md")
		if err := os.WriteFile(p, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
		mt := base.Add(time.Duration(i) * time.Hour)
		os.Chtimes(p, mt, mt)
	}
	return store
}

// press sends keys by name; a string that is not a key name is typed
// character by character. It returns the last message produced.
func press(m Model, keys ...string) (Model, tea.Msg) {
	var out tea.Msg
	for _, s := range keys {
		for _, key := range uitest.Type(s) {
			var cmd tea.Cmd
			m, cmd = m.Update(key)
			out = run(cmd)
		}
	}
	return m, out
}

func run(cmd tea.Cmd) tea.Msg {
	if cmd == nil {
		return nil
	}
	ch := make(chan tea.Msg, 1)
	go func() { ch <- cmd() }()
	select {
	case m := <-ch:
		switch m.(type) {
		case InsertMsg, CancelMsg, QuitMsg, tea.QuitMsg:
			return m
		}
		return nil // cursor blink and the like
	case <-time.After(20 * time.Millisecond):
		return nil
	}
}

func names(rows []notes.Note) []string {
	out := []string{}
	for _, r := range rows {
		out = append(out, r.Name)
	}
	return out
}

func TestPickerListsNewestFirst(t *testing.T) {
	m := New(setup(t, []string{"daily", "meeting"}, nil))
	if got := names(m.Rows()); !reflect.DeepEqual(got, []string{"meeting", "daily"}) {
		t.Errorf("rows = %v", got)
	}
	if m.Selected() != 0 || m.InForm() || m.Err() != "" {
		t.Errorf("sel=%d form=%v err=%q", m.Selected(), m.InForm(), m.Err())
	}
}

func TestPickerFilters(t *testing.T) {
	m := New(setup(t, []string{"daily", "meeting"}, nil))
	m, _ = press(m, "down")
	m, _ = press(m, "mt")
	if got := names(m.Rows()); !reflect.DeepEqual(got, []string{"meeting"}) {
		t.Errorf("rows = %v", got)
	}
	if m.Selected() != 0 {
		t.Errorf("selection not reset: %d", m.Selected())
	}
	if m.Query() != "mt" {
		t.Errorf("query = %q", m.Query())
	}
	m, _ = press(m, "backspace", "backspace")
	if len(m.Rows()) != 2 {
		t.Errorf("rows after clearing = %v", names(m.Rows()))
	}
}

func TestPickerNavigationNoWrap(t *testing.T) {
	m := New(setup(t, []string{"a", "b"}, nil))
	m, _ = press(m, "up")
	if m.Selected() != 0 {
		t.Errorf("sel after up at top = %d", m.Selected())
	}
	m, _ = press(m, "down", "down", "down")
	if m.Selected() != 1 {
		t.Errorf("sel = %d, want 1", m.Selected())
	}
	m, _ = press(m, "ctrl+p")
	if m.Selected() != 0 {
		t.Errorf("ctrl+p sel = %d", m.Selected())
	}
	m, _ = press(m, "ctrl+n")
	if m.Selected() != 1 {
		t.Errorf("ctrl+n sel = %d", m.Selected())
	}
}

func TestPickerEmptyHint(t *testing.T) {
	for _, store := range []notes.Store{setup(t, nil, nil), setup(t, []string{}, map[string]string{})} {
		m := New(store)
		if len(m.Rows()) != 0 || m.Err() != "" {
			t.Errorf("rows=%v err=%q", names(m.Rows()), m.Err())
		}
		if v := m.View(); !strings.Contains(v, "finder") || !strings.Contains(v, "ctrl+t") {
			t.Errorf("missing hint:\n%s", v)
		}
		if _, msg := press(m, "enter"); msg != nil {
			t.Errorf("enter on empty list -> %#v", msg)
		}
	}
	m := New(setup(t, []string{"a"}, nil))
	if strings.Contains(m.View(), "No templates yet") {
		t.Error("hint shown with templates present")
	}
}

func TestPickerEsc(t *testing.T) {
	m := New(setup(t, []string{"a"}, nil))
	if _, msg := press(m, "esc"); msg != (CancelMsg{}) {
		t.Errorf("esc -> %#v", msg)
	}
	if _, msg := press(m, "ctrl+c"); msg != tea.Msg(tea.QuitMsg{}) {
		t.Errorf("ctrl+c -> %#v", msg)
	}
}

func TestPickFieldlessInserts(t *testing.T) {
	m := New(setup(t, []string{"daily"}, map[string]string{"daily": "# Daily\n\n- [ ] \n"}))
	m, msg := press(m, "enter")
	if !reflect.DeepEqual(msg, InsertMsg{Text: "# Daily\n\n- [ ] "}) {
		t.Errorf("enter -> %#v", msg)
	}
	if m.InForm() {
		t.Error("form opened for a field-less template")
	}
}

func TestPickWithFieldsOpensForm(t *testing.T) {
	m := New(setup(t, []string{"meeting"}, map[string]string{"meeting": withFields}))
	m, msg := press(m, "enter")
	if msg != nil || !m.InForm() {
		t.Fatalf("enter -> %#v form=%v", msg, m.InForm())
	}
	if len(m.form) != 3 || m.Focus() != 0 {
		t.Errorf("rows=%d focus=%d", len(m.form), m.Focus())
	}
	want := map[string]string{"topic": "", "count": "0", "prio": "low"}
	if got := m.Values(); !reflect.DeepEqual(got, want) {
		t.Errorf("values = %v", got)
	}
}

func TestPickErrors(t *testing.T) {
	store := setup(t, []string{"gone", "bad"}, map[string]string{
		"bad": "---\nform:\n  - name: when\n    type: date\n---\n",
	})
	m := New(store) // bad is newer, listed first
	os.Remove(filepath.Join(store.Dir, "gone.md"))
	m, msg := press(m, "enter")
	if msg != nil || m.InForm() || !strings.Contains(m.Err(), "date") || !strings.Contains(m.View(), "date") {
		t.Errorf("bad template: msg=%#v form=%v err=%q", msg, m.InForm(), m.Err())
	}
	m, msg = press(m, "down")
	if m.Err() != "" {
		t.Error("error not cleared on key press")
	}
	m, msg = press(m, "enter")
	if msg != nil || m.InForm() || !strings.Contains(m.Err(), "cannot read") || !strings.Contains(m.Err(), "gone") {
		t.Errorf("missing template: msg=%#v form=%v err=%q", msg, m.InForm(), m.Err())
	}
}

func openForm(t *testing.T) Model {
	t.Helper()
	m := New(setup(t, []string{"meeting"}, map[string]string{"meeting": withFields}))
	m, _ = press(m, "enter")
	if !m.InForm() {
		t.Fatal("form did not open")
	}
	return m
}

func TestFormFocusMoves(t *testing.T) {
	m := openForm(t)
	m, msg := press(m, "enter")
	if msg != nil || m.Focus() != 1 {
		t.Errorf("enter: msg=%#v focus=%d", msg, m.Focus())
	}
	m, _ = press(m, "down")
	if m.Focus() != 2 {
		t.Errorf("down: focus=%d", m.Focus())
	}
	m, _ = press(m, "down")
	if m.Focus() != 2 {
		t.Errorf("down at last: focus=%d", m.Focus())
	}
	m, _ = press(m, "up")
	if m.Focus() != 1 {
		t.Errorf("up: focus=%d", m.Focus())
	}
	m, _ = press(m, "shift+tab")
	if m.Focus() != 0 {
		t.Errorf("shift+tab: focus=%d", m.Focus())
	}
	m, _ = press(m, "shift+tab")
	if m.Focus() != 0 {
		t.Errorf("shift+tab at first: focus=%d", m.Focus())
	}
	m, _ = press(m, "tab")
	if m.Focus() != 1 {
		t.Errorf("tab: focus=%d", m.Focus())
	}
	for i, r := range m.form {
		if r.input.Focused() != (i == m.Focus()) {
			t.Errorf("row %d focused=%v, focus=%d", i, r.input.Focused(), m.Focus())
		}
	}
}

func TestFormOptionCycling(t *testing.T) {
	m := openForm(t)
	m, _ = press(m, "tab", "tab") // prio row
	m, _ = press(m, "left")
	if got := m.Values()["prio"]; got != "low" {
		t.Errorf("left at first = %q", got)
	}
	m, _ = press(m, "right")
	if got := m.Values()["prio"]; got != "high" {
		t.Errorf("right = %q", got)
	}
	m, _ = press(m, "right")
	if got := m.Values()["prio"]; got != "high" {
		t.Errorf("right at last = %q", got)
	}
	m, _ = press(m, "x") // printable keys do nothing on an option row
	if got := m.Values()["prio"]; got != "high" {
		t.Errorf("after rune = %q", got)
	}
	m, _ = press(m, "left")
	if got := m.Values()["prio"]; got != "low" {
		t.Errorf("left = %q", got)
	}
}

func TestFormTextEditing(t *testing.T) {
	m := openForm(t)
	m, _ = press(m, "Standup")
	m, _ = press(m, "tab")
	m, _ = press(m, "backspace", "42")
	want := map[string]string{"topic": "Standup", "count": "42", "prio": "low"}
	if got := m.Values(); !reflect.DeepEqual(got, want) {
		t.Errorf("values = %v", got)
	}
	// Left/Right on a text row move the cursor instead of cycling.
	m, _ = press(m, "shift+tab", "left", "!")
	if got := m.Values()["topic"]; got != "Standu!p" {
		t.Errorf("topic = %q", got)
	}
}

func TestFormSubmit(t *testing.T) {
	m := openForm(t)
	m, _ = press(m, "Standup", "enter") // topic -> count
	m, _ = press(m, "enter")            // count -> prio
	m, _ = press(m, "right")            // high
	m, msg := press(m, "enter")         // submit
	if !reflect.DeepEqual(msg, InsertMsg{Text: "# Standup 0 high"}) {
		t.Errorf("submit -> %#v", msg)
	}
}

func TestFormInvalidNumber(t *testing.T) {
	m := openForm(t)
	m, _ = press(m, "tab", "backspace", "abc", "tab")
	m, msg := press(m, "enter")
	if msg != nil || !m.InForm() {
		t.Fatalf("submit with abc -> %#v form=%v", msg, m.InForm())
	}
	if !strings.Contains(m.Err(), "count") || !strings.Contains(m.View(), "count") {
		t.Errorf("err = %q", m.Err())
	}
	if m.Focus() != 1 || !m.form[1].input.Focused() || m.form[2].input.Focused() {
		t.Errorf("focus = %d", m.Focus())
	}
	// Fix it and submit from the last row.
	m, _ = press(m, "backspace", "backspace", "backspace", "7", "tab")
	m, msg = press(m, "enter")
	if !reflect.DeepEqual(msg, InsertMsg{Text: "#  7 low"}) {
		t.Errorf("resubmit -> %#v", msg)
	}
}

func TestFormDecimalVerbatim(t *testing.T) {
	m := openForm(t)
	m, _ = press(m, "tab", "backspace", "1.50", "tab")
	_, msg := press(m, "enter")
	if !reflect.DeepEqual(msg, InsertMsg{Text: "#  1.50 low"}) {
		t.Errorf("submit -> %#v", msg)
	}
}

func TestFormEsc(t *testing.T) {
	m := openForm(t)
	m, _ = press(m, "typed")
	if _, msg := press(m, "esc"); msg != (CancelMsg{}) {
		t.Errorf("esc -> %#v", msg)
	}
	if _, msg := press(m, "ctrl+c"); msg != tea.Msg(tea.QuitMsg{}) {
		t.Errorf("ctrl+c -> %#v", msg)
	}
}

func TestViewPicker(t *testing.T) {
	m := New(setup(t, []string{"daily", "meeting"}, nil))
	m, _ = press(m, "i")
	v := uitest.Strip(m.View())
	lines := strings.Split(v, "\n")
	if len(lines) != 24 {
		t.Errorf("height = %d, want 24", len(lines))
	}
	if !strings.HasPrefix(lines[0], "template> i") {
		t.Errorf("query line = %q", lines[0])
	}
	if lines[1] != "> meeting" || lines[2] != "  daily" {
		t.Errorf("rows = %q, %q", lines[1], lines[2])
	}
	if last := lines[len(lines)-1]; !strings.Contains(last, "enter") || !strings.Contains(last, "esc") || !strings.Contains(last, "ctrl+q quit") {
		t.Errorf("status bar = %q", last)
	}
	m, _ = m.Update(tea.WindowSizeMsg{Width: 40, Height: 5})
	lines = strings.Split(uitest.Strip(m.View()), "\n")
	if len(lines) != 5 {
		t.Errorf("height after resize = %d, want 5", len(lines))
	}
	if !strings.Contains(lines[1], "meeting") || strings.Contains(m.View(), "daily") {
		t.Errorf("window not shrunk:\n%s", uitest.Strip(m.View()))
	}
	m, _ = press(m, "down")
	if v := uitest.Strip(m.View()); !strings.Contains(v, "> daily") || strings.Contains(v, "meeting") {
		t.Errorf("window did not follow the selection:\n%s", v)
	}
}

func TestViewForm(t *testing.T) {
	m := openForm(t)
	m, _ = press(m, "Standup")
	lines := strings.Split(uitest.Strip(m.View()), "\n")
	if len(lines) != 24 {
		t.Errorf("height = %d, want 24", len(lines))
	}
	if !strings.HasPrefix(lines[0], "> Topic: Standup") {
		t.Errorf("row 0 = %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "  count: 0") {
		t.Errorf("row 1 = %q", lines[1])
	}
	if !strings.HasPrefix(lines[2], "  prio: < low > high") {
		t.Errorf("row 2 = %q", lines[2])
	}
	last := lines[len(lines)-1]
	for _, want := range []string{"enter", "tab", "esc", "ctrl+q quit"} {
		if !strings.Contains(last, want) {
			t.Errorf("status bar %q lacks %q", last, want)
		}
	}
	m, _ = press(m, "tab", "tab", "right")
	lines = strings.Split(uitest.Strip(m.View()), "\n")
	if !strings.HasPrefix(lines[2], "> prio: low < high >") {
		t.Errorf("row 2 after right = %q", lines[2])
	}
}

func TestCtrlQEmitsQuit(t *testing.T) {
	if _, msg := press(New(setup(t, []string{"a"}, nil)), "ctrl+q"); msg != (QuitMsg{}) {
		t.Errorf("picker ctrl+q -> %#v", msg)
	}
	m := openForm(t)
	m, _ = press(m, "typed")
	if _, msg := press(m, "ctrl+q"); msg != (QuitMsg{}) {
		t.Errorf("form ctrl+q -> %#v", msg)
	}
}
