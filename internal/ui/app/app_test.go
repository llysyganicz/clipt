package app

import (
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"

	"github.com/llysyganicz/clipt/internal/notes"
	"github.com/llysyganicz/clipt/internal/ui/finder"
	"github.com/llysyganicz/clipt/internal/ui/picker"
	"github.com/llysyganicz/clipt/internal/ui/theme"
	"github.com/llysyganicz/clipt/internal/ui/uitest"
)

func mkdir(t *testing.T, names ...string) string {
	t.Helper()
	dir := t.TempDir()
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i, n := range names { // later names are newer
		p := filepath.Join(dir, n+".md")
		if err := os.WriteFile(p, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		mt := base.Add(time.Duration(i) * time.Hour)
		os.Chtimes(p, mt, mt)
	}
	return dir
}

// feed delivers msg and then any message its command produces (one level).
func feed(m Model, msg tea.Msg) (Model, tea.Msg) {
	nm, cmd := m.Update(msg)
	m = nm.(Model)
	return m, run(cmd)
}

// press sends a key and re-delivers resulting app-level messages until quiet.
func press(m Model, keys ...string) (Model, bool) {
	quit := false
	for _, k := range keys {
		msg := tea.Msg(uitest.Key(k))
		for msg != nil {
			var out tea.Msg
			m, out = feed(m, msg)
			if _, ok := out.(tea.QuitMsg); ok {
				quit = true
			}
			msg = out
			if _, ok := msg.(tea.QuitMsg); ok {
				msg = nil
			}
		}
	}
	return m, quit
}

func TestSwitcherToggle(t *testing.T) {
	dir := mkdir(t, "c", "b", "a") // a newest
	m := New(dir, "")
	m, _ = press(m, "enter") // open a
	if m.mode != modePreview || m.current.Name != "a" {
		t.Fatalf("mode=%v current=%v", m.mode, m.current)
	}
	m, _ = press(m, "esc") // back to finder
	if m.mode != modeFinder {
		t.Fatalf("mode = %v", m.mode)
	}
	rows := m.finder.Rows()
	if rows[0].Note.Name != "b" {
		t.Errorf("top row = %q, want b", rows[0].Note.Name)
	}
	m, _ = press(m, "enter") // open b; a now next
	if m.current.Name != "b" {
		t.Errorf("current = %q", m.current.Name)
	}
}

func TestEscQuitsWithoutCurrent(t *testing.T) {
	m := New(mkdir(t, "a"), "")
	_, quit := press(m, "esc")
	if !quit {
		t.Error("esc without current note did not quit")
	}
}

func TestEscWithCurrentShowsPreview(t *testing.T) {
	m := New(mkdir(t, "a", "b"), "")
	m, _ = press(m, "enter")
	m, _ = press(m, "esc")     // note -> finder
	m, quit := press(m, "esc") // finder -> preview of current
	if quit || m.mode != modePreview || m.current == nil {
		t.Errorf("mode=%v quit=%v", m.mode, quit)
	}
}

func TestCurrentDisappeared(t *testing.T) {
	dir := mkdir(t, "a", "b")
	m := New(dir, "")
	m, _ = press(m, "enter") // open b
	m, _ = press(m, "esc")
	os.Remove(filepath.Join(dir, "b.md"))
	m, quit := press(m, "esc")
	if quit || m.mode != modeFinder || m.current != nil {
		t.Fatalf("mode=%v current=%v quit=%v", m.mode, m.current, quit)
	}
	if v := view(m); !strings.Contains(v, "no longer exists") {
		t.Errorf("view lacks error:\n%s", v)
	}
	if _, quit := press(m, "esc"); !quit {
		t.Error("esc after clearing current should quit")
	}
}

func TestCreateOpensNote(t *testing.T) {
	dir := mkdir(t)
	m := New(dir, "")
	m = typed(m, "new")
	m, _ = press(m, "enter")
	if m.mode != modeEdit || m.current == nil || m.current.Name != "new" {
		t.Errorf("mode=%v current=%v", m.mode, m.current)
	}
	if _, err := os.Stat(filepath.Join(dir, "new.md")); err != nil {
		t.Error(err)
	}
}

func TestPickerStartupCancelQuits(t *testing.T) {
	m := NewPicker("", "", "")
	_, quit := press(m, "esc")
	if !quit {
		t.Error("esc at startup picker did not quit")
	}
}

func TestPickerStartupAccept(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	dir := mkdir(t, "a")
	m, _ := feed(NewPicker("", "", ""), picker.DirAcceptedMsg{Dir: dir})
	if m.mode != modeFinder || m.dir != dir || len(m.finder.Rows()) != 1 {
		t.Errorf("mode=%v dir=%q rows=%d", m.mode, m.dir, len(m.finder.Rows()))
	}
}

func openPickerFromFinder(t *testing.T, m Model) Model {
	t.Helper()
	m, _ = press(m, "ctrl+o")
	if m.mode != modePicker || m.pickerFrom != picker.Finder {
		t.Fatalf("mode=%v from=%v", m.mode, m.pickerFrom)
	}
	return m
}

func TestPickerEscRestoresFinder(t *testing.T) {
	m := New(mkdir(t, "a", "b"), "")
	m, _ = press(m, "enter")
	m, _ = press(m, "esc")
	m = typed(m, "zz")
	cur := m.current
	m = openPickerFromFinder(t, m)
	m, quit := press(m, "esc")
	if quit || m.mode != modeFinder || m.current != cur || m.finder.Query() != "zz" {
		t.Errorf("mode=%v quit=%v query=%q", m.mode, quit, m.finder.Query())
	}
}

func TestDirChangeClearsCurrent(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	m := New(mkdir(t, "a", "b"), "")
	m, _ = press(m, "enter")
	m, _ = press(m, "esc")
	m = openPickerFromFinder(t, m)
	other := mkdir(t, "x")
	m, _ = feed(m, picker.DirAcceptedMsg{Dir: other})
	if m.current != nil || m.dir != other || m.finder.Query() != "" {
		t.Errorf("current=%v dir=%q query=%q", m.current, m.dir, m.finder.Query())
	}
	if _, quit := press(m, "esc"); !quit {
		t.Error("esc should quit after dir change")
	}
}

func TestSameDirKeepsCurrent(t *testing.T) {
	dir := mkdir(t, "a", "b")
	m := New(dir, "")
	m, _ = press(m, "enter")
	m, _ = press(m, "esc")
	m = openPickerFromFinder(t, m)
	m, _ = feed(m, picker.DirAcceptedMsg{Dir: dir})
	if m.current == nil || m.current.Name != "b" {
		t.Fatalf("current = %v", m.current)
	}
	for _, r := range m.finder.Rows() {
		if !r.IsCreate() && r.Note.Name == "b" {
			t.Error("current note should be hidden")
		}
	}
}

func TestOpenAndCreatedSetCurrent(t *testing.T) {
	m := New(mkdir(t, "o", "c"), "")
	m, _ = feed(m, finder.OpenNoteMsg{Note: notes.Note{Name: "o"}})
	if m.current.Name != "o" {
		t.Error("OpenNoteMsg")
	}
	m, _ = feed(m, finder.NoteCreatedMsg{Note: notes.Note{Name: "c"}})
	if m.current.Name != "c" {
		t.Error("NoteCreatedMsg")
	}
}

func TestNewPickerPrefill(t *testing.T) {
	if v := view(NewPicker("", "", "")); !strings.Contains(v, "~/notes") {
		t.Errorf("first run view lacks ~/notes:\n%s", v)
	}
	v := view(NewPicker("/mnt/gone", "directory does not exist: /mnt/gone", ""))
	if !strings.Contains(v, "/mnt/gone") || !strings.Contains(v, "does not exist") {
		t.Errorf("stale view:\n%s", v)
	}
}

// run executes a command, giving up quickly on blocking ones (cursor blink).
func run(cmd tea.Cmd) tea.Msg {
	if cmd == nil {
		return nil
	}
	ch := make(chan tea.Msg, 1)
	go func() { ch <- cmd() }()
	select {
	case m := <-ch:
		if _, ok := m.(tea.BatchMsg); ok {
			return nil
		}
		return m
	case <-time.After(20 * time.Millisecond):
		return nil
	}
}

// deliver sends msg and re-delivers resulting app-level messages until quiet.
func deliver(m Model, msg tea.Msg) (Model, bool) {
	quit := false
	for msg != nil {
		var out tea.Msg
		m, out = feed(m, msg)
		if _, ok := out.(tea.QuitMsg); ok {
			quit = true
			out = nil
		}
		msg = out
	}
	return m, quit
}

func rune1(r rune) tea.Msg { return uitest.Key(string(r)) }

// typed types s character by character without re-delivering any resulting
// message.
func typed(m Model, s string) Model {
	for _, k := range uitest.Type(s) {
		nm, _ := m.Update(k)
		m = nm.(Model)
	}
	return m
}

// view renders m without escape codes.
func view(m Model) string { return uitest.Strip(m.View().Content) }

func TestRenameUpdatesCurrent(t *testing.T) {
	dir := mkdir(t, "todo")
	m := New(dir, "")
	m, _ = press(m, "enter")
	m, _ = deliver(m, rune1('r'))
	m, _ = deliver(m, rune1('2'))
	m, _ = deliver(m, uitest.Key("enter"))
	if m.mode != modePreview || m.current == nil || m.current.Name != "todo2" {
		t.Fatalf("mode=%v current=%v", m.mode, m.current)
	}
	if _, err := os.Stat(filepath.Join(dir, "todo2.md")); err != nil {
		t.Error(err)
	}
	m, _ = press(m, "esc") // back to finder, renamed note hidden
	if m.mode != modeFinder || m.finder.Query() != "" {
		t.Fatalf("mode=%v query=%q", m.mode, m.finder.Query())
	}
	for _, r := range m.finder.Rows() {
		if !r.IsCreate() && r.Note.Name == "todo2" {
			t.Error("current note listed")
		}
	}
}

func TestDeleteClearsCurrent(t *testing.T) {
	dir := mkdir(t, "keep", "todo") // todo newest
	m := New(dir, "")
	m, _ = press(m, "enter")
	m, _ = deliver(m, rune1('d'))
	m, _ = deliver(m, rune1('y'))
	if m.mode != modeFinder || m.current != nil || m.finder.Query() != "" {
		t.Fatalf("mode=%v current=%v", m.mode, m.current)
	}
	if _, err := os.Stat(filepath.Join(dir, "todo.md")); !os.IsNotExist(err) {
		t.Error("todo.md still exists")
	}
	rows := m.finder.Rows()
	if len(rows) != 1 || rows[0].Note.Name != "keep" {
		t.Errorf("rows = %v", rows)
	}
	if _, quit := press(m, "esc"); !quit {
		t.Error("esc should quit with no current note")
	}
}

func TestPreviewBackEmptyQuery(t *testing.T) {
	m := New(mkdir(t, "a", "b"), "")
	m, _ = press(m, "enter")
	m, _ = press(m, "esc") // finder
	nm, _ := m.Update(rune1('x'))
	m = nm.(Model)
	m, _ = press(m, "esc") // preview of current
	m, _ = press(m, "esc") // finder again
	if m.mode != modeFinder || m.finder.Query() != "" || m.current == nil {
		t.Errorf("mode=%v query=%q current=%v", m.mode, m.finder.Query(), m.current)
	}
}

func TestUnreadableNoteReturnsToFinder(t *testing.T) {
	dir := mkdir(t, "a", "b")
	m := New(dir, "")
	os.Remove(filepath.Join(dir, "b.md")) // listed, then removed
	m, _ = press(m, "enter")
	if m.mode != modeFinder || m.current != nil {
		t.Fatalf("mode=%v current=%v", m.mode, m.current)
	}
	if v := view(m); !strings.Contains(v, "cannot read") || !strings.Contains(v, "b") {
		t.Errorf("view lacks error:\n%s", v)
	}
}

func TestEditReturnsToPreview(t *testing.T) {
	m := New(mkdir(t, "a"), "")
	m, _ = press(m, "enter")
	m, _ = deliver(m, rune1('e'))
	if m.mode != modeEdit {
		t.Fatalf("mode=%v", m.mode)
	}
	if v := view(m); !strings.Contains(v, "esc save & back") {
		t.Errorf("editor view lacks hint:\n%s", v)
	}
	m, _ = press(m, "esc")
	if m.mode != modePreview || m.current == nil || m.current.Name != "a" {
		t.Errorf("mode=%v current=%v", m.mode, m.current)
	}
}

func TestCreateTypeEscShowsPreview(t *testing.T) {
	dir := mkdir(t)
	m := New(dir, "")
	m = typed(m, "ideas")
	m, _ = press(m, "enter")
	if m.mode != modeEdit {
		t.Fatalf("mode=%v", m.mode)
	}
	m = typed(m, "# Brainstorm")
	m, _ = press(m, "esc")
	if m.mode != modePreview || m.current == nil || m.current.Name != "ideas" {
		t.Fatalf("mode=%v current=%v", m.mode, m.current)
	}
	if v := view(m); !strings.Contains(v, "Brainstorm") || !strings.Contains(v, "ideas") {
		t.Errorf("preview lacks typed content or name:\n%s", v)
	}
	if b, err := os.ReadFile(filepath.Join(dir, "ideas.md")); err != nil || string(b) != "# Brainstorm\n" {
		t.Errorf("ideas.md = %q, %v", b, err)
	}
}

func TestEditUnreadableReturnsToFinder(t *testing.T) {
	dir := mkdir(t, "a")
	m := New(dir, "")
	m, _ = press(m, "enter")
	os.Remove(filepath.Join(dir, "a.md"))
	m, _ = deliver(m, rune1('e'))
	if m.mode != modeFinder || m.current != nil {
		t.Fatalf("mode=%v current=%v", m.mode, m.current)
	}
	if v := view(m); !strings.Contains(v, "cannot read") {
		t.Errorf("view lacks error:\n%s", v)
	}
}

func TestResizeReachesEditor(t *testing.T) {
	m := New(mkdir(t, "a"), "")
	m, _ = feed(m, tea.WindowSizeMsg{Width: 60, Height: 12})
	m, _ = press(m, "enter")
	m, _ = deliver(m, rune1('e'))
	if got := strings.Count(view(m), "\n") + 1; got != 12 {
		t.Errorf("editor height = %d, want 12", got)
	}
	m, _ = feed(m, tea.WindowSizeMsg{Width: 60, Height: 8})
	if got := strings.Count(view(m), "\n") + 1; got != 8 {
		t.Errorf("after resize height = %d, want 8", got)
	}
}

func TestResizeReachesPreview(t *testing.T) {
	m := New(mkdir(t, "a"), "")
	m, _ = feed(m, tea.WindowSizeMsg{Width: 60, Height: 12})
	m, _ = press(m, "enter")
	if got := strings.Count(view(m), "\n") + 1; got != 12 {
		t.Errorf("preview height = %d, want 12", got)
	}
	m, _ = feed(m, tea.WindowSizeMsg{Width: 60, Height: 8})
	if got := strings.Count(view(m), "\n") + 1; got != 8 {
		t.Errorf("after resize height = %d, want 8", got)
	}
}

// addTemplates writes templates under dir/.templates (later names are newer).
func addTemplates(t *testing.T, dir string, files map[string]string, order ...string) {
	t.Helper()
	tdir := filepath.Join(dir, ".templates")
	if err := os.MkdirAll(tdir, 0o755); err != nil {
		t.Fatal(err)
	}
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i, n := range order {
		p := filepath.Join(tdir, n+".md")
		if err := os.WriteFile(p, []byte(files[n]), 0o644); err != nil {
			t.Fatal(err)
		}
		mt := base.Add(time.Duration(i) * time.Hour)
		os.Chtimes(p, mt, mt)
	}
}

func typeRunes(m Model, s string) Model {
	for _, r := range s {
		m, _ = deliver(m, rune1(r))
	}
	return m
}

func TestFinderScopeToggle(t *testing.T) {
	dir := mkdir(t, "todo")
	addTemplates(t, dir, map[string]string{"meeting": "# m\n"}, "meeting")
	m := New(dir, "")
	m = typeRunes(m, "to")
	m, _ = press(m, "ctrl+t")
	if m.mode != modeFinder || m.finder.Scope() != finder.ScopeTemplates || m.finder.Query() != "" {
		t.Fatalf("mode=%v scope=%v query=%q", m.mode, m.finder.Scope(), m.finder.Query())
	}
	rows := m.finder.Rows()
	if len(rows) != 1 || rows[0].Note.Name != "meeting" {
		t.Errorf("template rows = %v", rows)
	}
	if v := view(m); !strings.Contains(v, "templates> ") {
		t.Errorf("view lacks the template prompt:\n%s", v)
	}
	m, _ = press(m, "ctrl+t")
	if m.finder.Scope() != finder.ScopeNotes || m.finder.Query() != "" {
		t.Errorf("scope=%v query=%q after toggling back", m.finder.Scope(), m.finder.Query())
	}
	if rows := m.finder.Rows(); len(rows) != 1 || rows[0].Note.Name != "todo" {
		t.Errorf("note rows = %v", rows)
	}
}

func TestTemplateScopeEscReturnsToNotes(t *testing.T) {
	m := New(mkdir(t, "todo"), "")
	m, _ = press(m, "ctrl+t")
	m, quit := press(m, "esc")
	if quit || m.mode != modeFinder || m.finder.Scope() != finder.ScopeNotes {
		t.Errorf("quit=%v mode=%v scope=%v", quit, m.mode, m.finder.Scope())
	}
}

func TestEditTemplateKeepsCurrent(t *testing.T) {
	dir := mkdir(t, "todo")
	addTemplates(t, dir, map[string]string{"meeting": "# m\n"}, "meeting")
	m := New(dir, "")
	m, _ = press(m, "enter") // todo is current
	m, _ = press(m, "esc")
	m, _ = press(m, "ctrl+t")
	m, _ = press(m, "enter") // open meeting
	if m.mode != modeEdit || m.current == nil || m.current.Name != "todo" {
		t.Fatalf("mode=%v current=%v", m.mode, m.current)
	}
	if v := view(m); !strings.Contains(v, "template: meeting") || strings.Contains(v, "ctrl+t") {
		t.Errorf("editor view:\n%s", v)
	}
	m, _ = press(m, "ctrl+t") // nothing on a template
	if m.mode != modeEdit {
		t.Errorf("ctrl+t on a template changed mode to %v", m.mode)
	}
	m = typeRunes(m, "\n- a")
	m, _ = press(m, "esc")
	if m.mode != modeFinder || m.finder.Scope() != finder.ScopeTemplates || m.finder.Query() != "" {
		t.Fatalf("mode=%v scope=%v query=%q", m.mode, m.finder.Scope(), m.finder.Query())
	}
	if m.current == nil || m.current.Name != "todo" {
		t.Errorf("current = %v", m.current)
	}
	b, err := os.ReadFile(filepath.Join(dir, ".templates", "meeting.md"))
	if err != nil || string(b) != "# m\n- a\n" {
		t.Errorf("template file = %q, %v", b, err)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "todo.md")); len(b) != 0 {
		t.Errorf("note touched: %q", b)
	}
}

func TestCreateTemplateFromFinder(t *testing.T) {
	dir := mkdir(t, "todo")
	m := New(dir, "")
	m, _ = press(m, "ctrl+t")
	m = typeRunes(m, "meeting")
	m, _ = press(m, "enter")
	if m.mode != modeEdit || m.current != nil {
		t.Fatalf("mode=%v current=%v", m.mode, m.current)
	}
	if v := view(m); !strings.Contains(v, "{{title}}") || !strings.Contains(v, "template: meeting") {
		t.Errorf("editor view:\n%s", v)
	}
	m, _ = press(m, "esc")
	if m.mode != modeFinder || m.finder.Scope() != finder.ScopeTemplates {
		t.Fatalf("mode=%v scope=%v", m.mode, m.finder.Scope())
	}
	if rows := m.finder.Rows(); len(rows) != 1 || rows[0].Note.Name != "meeting" {
		t.Errorf("template rows = %v", rows)
	}
	m, _ = press(m, "ctrl+t")
	if rows := m.finder.Rows(); len(rows) != 1 || rows[0].Note.Name != "todo" {
		t.Errorf(".templates leaked into notes: %v", rows)
	}
}

func TestChangeDirFromTemplateScope(t *testing.T) {
	m := New(mkdir(t, "todo"), "")
	m, _ = press(m, "ctrl+t")
	m = openPickerFromFinder(t, m)
	m, quit := press(m, "esc")
	if quit || m.mode != modeFinder || m.finder.Scope() != finder.ScopeNotes {
		t.Errorf("quit=%v mode=%v scope=%v", quit, m.mode, m.finder.Scope())
	}
}

func TestUnreadableTemplateKeepsScopeAndCurrent(t *testing.T) {
	dir := mkdir(t, "todo")
	addTemplates(t, dir, map[string]string{"meeting": "# m\n"}, "meeting")
	m := New(dir, "")
	m, _ = press(m, "enter")
	m, _ = press(m, "esc")
	m, _ = press(m, "ctrl+t")
	os.Remove(filepath.Join(dir, ".templates", "meeting.md"))
	m, _ = press(m, "enter")
	if m.mode != modeFinder || m.finder.Scope() != finder.ScopeTemplates {
		t.Fatalf("mode=%v scope=%v", m.mode, m.finder.Scope())
	}
	if m.current == nil || m.current.Name != "todo" {
		t.Errorf("current = %v", m.current)
	}
	if v := view(m); !strings.Contains(v, "cannot read") || !strings.Contains(v, "meeting") {
		t.Errorf("view lacks error:\n%s", v)
	}
}

// openNoteForInsert opens the note "n" (content "ab") in edit mode with the
// cursor between a and b and the given templates present.
func openNoteForInsert(t *testing.T, files map[string]string, order ...string) (Model, string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "n.md"), []byte("ab\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	addTemplates(t, dir, files, order...)
	m := New(dir, "")
	m, _ = feed(m, tea.WindowSizeMsg{Width: 60, Height: 12})
	m, _ = press(m, "enter")
	m, _ = deliver(m, rune1('e'))
	m, _ = press(m, "left")
	if m.mode != modeEdit {
		t.Fatalf("mode = %v", m.mode)
	}
	return m, dir
}

func TestInsertTemplateWithForm(t *testing.T) {
	tpl := "---\nform:\n  - name: topic\n    type: text\n  - name: prio\n    type: option\n    options: [low, high]\n---\n# {{topic}} ({{prio}})\n- "
	m, dir := openNoteForInsert(t, map[string]string{"meeting": tpl, "daily": "x"}, "daily", "meeting")
	m, _ = press(m, "ctrl+t")
	if m.mode != modeTemplate {
		t.Fatalf("mode = %v", m.mode)
	}
	if v := view(m); !strings.Contains(v, "meeting") || !strings.Contains(v, "daily") {
		t.Errorf("picker view:\n%s", v)
	}
	if got := strings.Count(view(m), "\n") + 1; got != 12 {
		t.Errorf("picker height = %d, want 12", got)
	}
	m, _ = press(m, "enter") // meeting (newest) opens the form
	if m.mode != modeTemplate || !m.tmpl.InForm() {
		t.Fatalf("mode=%v form=%v", m.mode, m.tmpl.InForm())
	}
	m = typeRunes(m, "Standup")
	m, _ = press(m, "enter", "right", "enter")
	if m.mode != modeEdit {
		t.Fatalf("mode after submit = %v", m.mode)
	}
	if v := view(m); !strings.Contains(v, "# Standup (high)") || !strings.Contains(v, "esc save & back") {
		t.Errorf("editor view:\n%s", v)
	}
	m = typeRunes(m, "X") // cursor is right after the inserted "- "
	m, _ = press(m, "esc")
	if m.mode != modePreview {
		t.Errorf("mode after esc = %v", m.mode)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "n.md"))
	if string(b) != "a# Standup (high)\n- Xb\n" {
		t.Errorf("n.md = %q", b)
	}
}

func TestInsertFieldlessTemplate(t *testing.T) {
	m, dir := openNoteForInsert(t, map[string]string{"daily": "# Daily\n\n- [ ] \n"}, "daily")
	m, _ = press(m, "ctrl+t", "enter")
	if m.mode != modeEdit {
		t.Fatalf("mode = %v", m.mode)
	}
	m, _ = press(m, "esc")
	b, _ := os.ReadFile(filepath.Join(dir, "n.md"))
	if string(b) != "a# Daily\n\n- [ ] b\n" {
		t.Errorf("n.md = %q", b)
	}
}

func TestInsertCancelLeavesEditorUntouched(t *testing.T) {
	for _, enter := range []bool{false, true} {
		tpl := "---\nform:\n  - name: topic\n    type: text\n---\n# {{topic}}\n"
		m, dir := openNoteForInsert(t, map[string]string{"meeting": tpl}, "meeting")
		m, _ = press(m, "ctrl+t")
		if enter {
			m, _ = press(m, "enter")
			m = typeRunes(m, "typed")
		} else {
			m = typeRunes(m, "me")
		}
		m, quit := press(m, "esc")
		if quit || m.mode != modeEdit {
			t.Fatalf("enter=%v: quit=%v mode=%v", enter, quit, m.mode)
		}
		if v := view(m); !strings.Contains(v, "esc save & back") || strings.Contains(v, "typed") {
			t.Errorf("editor view:\n%s", v)
		}
		m = typeRunes(m, "X") // cursor still between a and b
		m, _ = press(m, "esc")
		b, _ := os.ReadFile(filepath.Join(dir, "n.md"))
		if string(b) != "aXb\n" {
			t.Errorf("enter=%v: n.md = %q", enter, b)
		}
	}
}

func TestInsertCancelWithoutTypingDoesNotWrite(t *testing.T) {
	m, dir := openNoteForInsert(t, map[string]string{"meeting": "x"}, "meeting")
	p := filepath.Join(dir, "n.md")
	old := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	os.Chtimes(p, old, old)
	m, _ = press(m, "ctrl+t", "esc", "esc")
	if m.mode != modePreview {
		t.Fatalf("mode = %v", m.mode)
	}
	if info, _ := os.Stat(p); !info.ModTime().Equal(old) {
		t.Error("note rewritten after a cancelled insertion")
	}
}

func TestInsertModeResize(t *testing.T) {
	m, _ := openNoteForInsert(t, map[string]string{"meeting": "x"}, "meeting")
	m, _ = press(m, "ctrl+t")
	m, _ = feed(m, tea.WindowSizeMsg{Width: 60, Height: 8})
	if got := strings.Count(view(m), "\n") + 1; got != 8 {
		t.Errorf("picker height after resize = %d, want 8", got)
	}
	m, _ = press(m, "esc")
	if got := strings.Count(view(m), "\n") + 1; got != 8 {
		t.Errorf("editor height after resize = %d, want 8", got)
	}
}

func TestCtrlQFromTemplateFormSavesNoteAndQuits(t *testing.T) {
	tpl := "---\nform:\n  - name: topic\n    type: text\n---\n# {{topic}}\n"
	m, dir := openNoteForInsert(t, map[string]string{"meeting": tpl}, "meeting")
	m = typeRunes(m, "X") // note is now aXb
	m, _ = press(m, "ctrl+t", "enter")
	if m.mode != modeTemplate || !m.tmpl.InForm() {
		t.Fatalf("mode=%v form=%v", m.mode, m.tmpl.InForm())
	}
	m = typeRunes(m, "typed")
	m, quit := press(m, "ctrl+q")
	if !quit {
		t.Error("ctrl+q from the form did not quit")
	}
	b, _ := os.ReadFile(filepath.Join(dir, "n.md"))
	if string(b) != "aXb\n" {
		t.Errorf("n.md = %q", b)
	}
}

func TestCtrlQFromTemplatePickerWithFailingWrite(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores permissions")
	}
	m, dir := openNoteForInsert(t, map[string]string{"meeting": "x"}, "meeting")
	m = typeRunes(m, "X")
	m, _ = press(m, "ctrl+t")
	p := filepath.Join(dir, "n.md")
	if err := os.Chmod(p, 0o444); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0o755); os.Chmod(p, 0o644) })
	m, quit := press(m, "ctrl+q")
	if quit || m.mode != modeEdit {
		t.Fatalf("quit=%v mode=%v", quit, m.mode)
	}
	if v := view(m); !strings.Contains(v, "permission denied") || !strings.Contains(v, "aXb") {
		t.Errorf("editor view:\n%s", v)
	}
	b, _ := os.ReadFile(p)
	if string(b) != "ab\n" {
		t.Errorf("n.md = %q", b)
	}
}

func TestCtrlQQuitsFromFinderPreviewAndEditor(t *testing.T) {
	dir := mkdir(t, "a")
	m := New(dir, "")
	if _, quit := press(m, "ctrl+q"); !quit {
		t.Error("finder: ctrl+q did not quit")
	}
	m, _ = press(m, "enter")
	if _, quit := press(m, "ctrl+q"); !quit {
		t.Error("preview: ctrl+q did not quit")
	}
	m, _ = deliver(m, rune1('e'))
	m = typeRunes(m, "hi")
	if _, quit := press(m, "ctrl+q"); !quit {
		t.Error("editor: ctrl+q did not quit")
	}
	b, _ := os.ReadFile(filepath.Join(dir, "a.md"))
	if string(b) != "hi\n" {
		t.Errorf("a.md = %q", b)
	}
}

func TestViewUsesAltScreen(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.md"), []byte("# A\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m := New(dir, "")
	if !m.View().AltScreen {
		t.Error("finder view is not in the alt screen")
	}
	m, _ = press(m, "enter")
	if m.mode != modePreview || !m.View().AltScreen {
		t.Errorf("preview: mode=%v alt=%v", m.mode, m.View().AltScreen)
	}
	m, _ = deliver(m, rune1('e'))
	if m.mode != modeEdit || !m.View().AltScreen {
		t.Errorf("edit: mode=%v alt=%v", m.mode, m.View().AltScreen)
	}
	if !strings.Contains(view(m), "# A") {
		t.Errorf("edit view lacks the note:\n%s", view(m))
	}
}

// withPaintPerCell sets the root view's mechanism for the test.
func withPaintPerCell(t *testing.T, on bool) {
	t.Helper()
	was := paintPerCell
	t.Cleanup(func() { paintPerCell = was })
	paintPerCell = on
}

// withLightTheme applies gruvbox_light for the test and restores gruvbox_dark
// after it.
func withLightTheme(t *testing.T) theme.Palette {
	t.Helper()
	light, err := theme.Load("gruvbox_light", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	dark, err := theme.Load("gruvbox_dark", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { theme.Apply(dark) })
	theme.Apply(light)
	return light
}

func TestViewPaintsThemeColors(t *testing.T) {
	withPaintPerCell(t, false)
	light, err := theme.Load("gruvbox_light", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	dark, err := theme.Load("gruvbox_dark", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { theme.Apply(dark) })
	theme.Apply(light)

	dir := mkdir(t, "a")
	addTemplates(t, dir, map[string]string{"meeting": "# m\n"}, "meeting")
	check := func(what string, m Model, want mode) {
		t.Helper()
		v := m.View()
		if m.mode != want {
			t.Fatalf("%s: mode = %v, want %v", what, m.mode, want)
		}
		if v.BackgroundColor != lipgloss.Color(light.Background) {
			t.Errorf("%s: background = %v, want %s", what, v.BackgroundColor, light.Background)
		}
		if v.ForegroundColor != lipgloss.Color(light.Foreground) {
			t.Errorf("%s: foreground = %v, want %s", what, v.ForegroundColor, light.Foreground)
		}
	}
	m := New(dir, "")
	check("finder", m, modeFinder)
	m = openPickerFromFinder(t, m)
	check("picker", m, modePicker)
	m, _ = press(m, "esc")
	check("finder again", m, modeFinder)
	m, _ = press(m, "enter")
	check("preview", m, modePreview)
	m, _ = deliver(m, rune1('e'))
	check("edit", m, modeEdit)
	m, _ = press(m, "ctrl+t")
	check("template picker", m, modeTemplate)
}

func TestCreateWithTagsStartsAfterFrontMatter(t *testing.T) {
	const created = "---\ntags: [work, meeting]\n---\n"
	dir := mkdir(t)
	m := New(dir, "")
	m = typed(m, "#work #meeting standup")
	m, _ = press(m, "enter")
	if m.mode != modeEdit || m.current == nil || m.current.Name != "standup" {
		t.Fatalf("mode=%v current=%v", m.mode, m.current)
	}
	p := filepath.Join(dir, "standup.md")
	old := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	os.Chtimes(p, old, old)
	m, _ = press(m, "esc") // untouched: nothing is rewritten
	info, _ := os.Stat(p)
	if b, _ := os.ReadFile(p); string(b) != created || !info.ModTime().Equal(old) {
		t.Errorf("untouched save: %q, rewritten=%v", b, !info.ModTime().Equal(old))
	}

	// The cursor row is the line after the fence right after creation.
	m = New(dir, "")
	m = typed(m, "#work review")
	m, _ = press(m, "down", "enter")
	m = typed(m, "x")
	m, _ = press(m, "esc")
	if b, _ := os.ReadFile(filepath.Join(dir, "review.md")); string(b) != "---\ntags: [work]\n---\nx\n" {
		t.Errorf("review.md = %q", b)
	}
}

func TestInsertTaggedTemplateMergesTags(t *testing.T) {
	m, dir := openNoteForInsert(t, map[string]string{"meeting": "---\ntags: [meeting]\n---\n- agenda\n"}, "meeting")
	m, _ = press(m, "ctrl+t", "enter")
	if m.mode != modeEdit {
		t.Fatalf("mode = %v", m.mode)
	}
	m = typeRunes(m, "X")
	m, _ = press(m, "esc")
	b, _ := os.ReadFile(filepath.Join(dir, "n.md"))
	if string(b) != "---\ntags: [meeting]\n---\na- agendaXb\n" {
		t.Errorf("n.md = %q", b)
	}
}

func TestInsertTaggedTemplateIntoBrokenFrontMatter(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "n.md"), []byte("---\ntags: [work\n---\nab\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	addTemplates(t, dir, map[string]string{"meeting": "---\ntags: [meeting]\n---\n!\n"}, "meeting")
	m := New(dir, "")
	m, _ = feed(m, tea.WindowSizeMsg{Width: 60, Height: 12})
	m, _ = press(m, "enter")
	m, _ = deliver(m, rune1('e'))
	m, _ = press(m, "ctrl+t", "enter")
	if m.mode != modeEdit {
		t.Fatalf("mode = %v", m.mode)
	}
	if v := view(m); !strings.Contains(v, "tags not merged") {
		t.Errorf("no error in editor:\n%s", v)
	}
	m, _ = press(m, "esc")
	b, _ := os.ReadFile(filepath.Join(dir, "n.md"))
	if string(b) != "---\ntags: [work\n---\nab!\n" {
		t.Errorf("n.md = %q", b)
	}
}

func TestDirChangeKeepsTheme(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	m := New(mkdir(t, "a"), "gruvbox_light")
	m = openPickerFromFinder(t, m)
	other := mkdir(t, "x")
	m, _ = press(m, "ctrl+u") // clear the prefilled current directory
	m = typed(m, other)
	m, _ = press(m, "enter")
	if m.mode != modeFinder || m.dir != other {
		t.Fatalf("mode=%v dir=%q", m.mode, m.dir)
	}
	b, err := os.ReadFile(filepath.Join(xdg, "clipt", "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `theme = "gruvbox_light"`) || !strings.Contains(string(b), other) {
		t.Errorf("rewritten config:\n%s", b)
	}
}

func TestDirChangeWithoutThemeWritesNone(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	m := New(mkdir(t, "a"), "")
	m = openPickerFromFinder(t, m)
	m, _ = press(m, "ctrl+u")
	m = typed(m, mkdir(t, "x"))
	m, _ = press(m, "enter")
	b, err := os.ReadFile(filepath.Join(xdg, "clipt", "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "theme") {
		t.Errorf("config without a theme gained one:\n%s", b)
	}
}

// sameColor reports whether two colors have the same RGBA values.
func sameColor(a, b color.Color) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return ar == br && ag == bg && ab == bb && aa == ba
}

func TestViewPaintsPerCell(t *testing.T) {
	withPaintPerCell(t, true)
	light := withLightTheme(t)
	const width, height = 60, 12

	dir := mkdir(t, "a")
	addTemplates(t, dir, map[string]string{"meeting": "# m\n"}, "meeting")
	check := func(what string, m Model, want mode) {
		t.Helper()
		if m.mode != want {
			t.Fatalf("%s: mode = %v, want %v", what, m.mode, want)
		}
		v := m.View()
		if v.BackgroundColor != nil {
			t.Errorf("%s: view sets background %v", what, v.BackgroundColor)
		}
		if v.ForegroundColor != nil {
			t.Errorf("%s: view sets foreground %v", what, v.ForegroundColor)
		}
		if !v.AltScreen {
			t.Errorf("%s: view is not in the alt screen", what)
		}
		first, _, _ := strings.Cut(v.Content, "\n")
		line := uv.NewStyledString(first).Lines(ansi.WcWidth)[0]
		if len(line) != width {
			t.Fatalf("%s: first line has %d cells, want %d:\n%q", what, len(line), width, first)
		}
		for x := range line {
			if c := line.At(x); !sameColor(c.Style.Bg, lipgloss.Color(light.Background)) {
				t.Errorf("%s: cell %d background = %v, want %s", what, x, c.Style.Bg, light.Background)
			}
		}
		if got := strings.Count(v.Content, "\n") + 1; got != height {
			t.Errorf("%s: view has %d lines, want %d", what, got, height)
		}
	}
	m := New(dir, "")
	m, _ = feed(m, tea.WindowSizeMsg{Width: width, Height: height})
	check("finder", m, modeFinder)
	m = openPickerFromFinder(t, m)
	check("picker", m, modePicker)
	m, _ = press(m, "esc")
	check("finder again", m, modeFinder)
	m, _ = press(m, "enter")
	check("preview", m, modePreview)
	m, _ = deliver(m, rune1('e'))
	check("edit", m, modeEdit)
	m, _ = press(m, "ctrl+t")
	check("template picker", m, modeTemplate)
}

func TestViewPerCellBeforeSizeIsUnpadded(t *testing.T) {
	withPaintPerCell(t, true)
	withLightTheme(t)
	m := New(mkdir(t, "a"), "")
	v := m.View()
	if v.Content != m.content() {
		t.Errorf("content before any window size was changed:\n%q\nwant\n%q", v.Content, m.content())
	}
	if v.BackgroundColor != nil || v.ForegroundColor != nil {
		t.Errorf("view sets terminal colors: bg=%v fg=%v", v.BackgroundColor, v.ForegroundColor)
	}
}
