package finder

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/llysyganicz/clipt/internal/notes"
	"github.com/llysyganicz/clipt/internal/templates"
)

func newDir(t *testing.T, names ...string) notes.Store {
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
	return notes.Store{Dir: dir}
}

func k(s string) tea.KeyMsg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "ctrl+p":
		return tea.KeyMsg{Type: tea.KeyCtrlP}
	case "ctrl+n":
		return tea.KeyMsg{Type: tea.KeyCtrlN}
	case "ctrl+o":
		return tea.KeyMsg{Type: tea.KeyCtrlO}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case "ctrl+t":
		return tea.KeyMsg{Type: tea.KeyCtrlT}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func press(m Model, keys ...string) (Model, tea.Msg) {
	var out tea.Msg
	for _, s := range keys {
		nm, cmd := m.Update(k(s))
		m = nm.(Model)
		out = run(cmd)
	}
	return m, out
}

func TestNavigationNoWrap(t *testing.T) {
	m := New(newDir(t, "a", "b"), nil, "")
	m, _ = press(m, "up")
	if m.Selected() != 0 {
		t.Errorf("sel = %d after up at top", m.Selected())
	}
	m, _ = press(m, "down", "down", "down")
	if m.Selected() != 1 {
		t.Errorf("sel = %d, want 1 (no wrap)", m.Selected())
	}
	m, _ = press(m, "ctrl+p")
	if m.Selected() != 0 {
		t.Errorf("ctrl+p: sel = %d", m.Selected())
	}
	m, _ = press(m, "ctrl+n")
	if m.Selected() != 1 {
		t.Errorf("ctrl+n: sel = %d", m.Selected())
	}
}

func TestDownToCreateRow(t *testing.T) {
	m := New(newDir(t, "todos", "todo-b"), nil, "")
	m, _ = press(m, "t", "o", "d", "o")
	m, _ = press(m, "down", "down")
	if r := m.Rows()[m.Selected()]; !r.IsCreate() || r.Create != "todo" {
		t.Errorf("selected = %+v, want create row", r)
	}
}

func TestSelectionResetsOnQueryChange(t *testing.T) {
	m := New(newDir(t, "aa", "ab", "ac"), nil, "")
	m, _ = press(m, "down", "down")
	m, _ = press(m, "a")
	if m.Selected() != 0 {
		t.Errorf("sel = %d, want 0", m.Selected())
	}
}

func TestKeys(t *testing.T) {
	m := New(newDir(t, "a"), nil, "xy")
	m, msg := press(m, "esc")
	if _, ok := msg.(BackMsg); !ok {
		t.Errorf("esc msg = %#v", msg)
	}
	if m.Query() != "xy" {
		t.Errorf("esc cleared query: %q", m.Query())
	}
	if _, msg := press(m, "ctrl+o"); msg != (ChangeDirMsg{}) {
		t.Errorf("ctrl+o msg = %#v", msg)
	}
	_, msg = press(m, "ctrl+c")
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("ctrl+c msg = %#v", msg)
	}
}

func TestEnterOpensNote(t *testing.T) {
	m := New(newDir(t, "old", "new"), nil, "")
	_, msg := press(m, "enter")
	if got, ok := msg.(OpenNoteMsg); !ok || got.Note.Name != "new" || got.Scope != ScopeNotes {
		t.Errorf("msg = %#v", msg)
	}
}

func TestEnterOnEmptyListIsNoop(t *testing.T) {
	m := New(newDir(t), nil, "")
	_, msg := press(m, "enter")
	if msg != nil {
		t.Errorf("msg = %#v", msg)
	}
}

func TestEnterCreates(t *testing.T) {
	store := newDir(t, "x")
	m := New(store, nil, "foo.md")
	_, msg := press(m, "enter") // x does not match, so only the create row
	got, ok := msg.(NoteCreatedMsg)
	if !ok || got.Note.Name != "foo.md" || got.Scope != ScopeNotes {
		t.Fatalf("msg = %#v", msg)
	}
	if _, err := os.Stat(filepath.Join(store.Dir, "foo.md.md")); err != nil {
		t.Errorf("file not created: %v", err)
	}
}

func TestCreateInvalidName(t *testing.T) {
	store := newDir(t)
	m := New(store, nil, "a/b")
	m, msg := press(m, "enter")
	if msg != nil || m.err == "" || !strings.Contains(m.View(), m.err) {
		t.Errorf("msg = %#v, err = %q", msg, m.err)
	}
	entries, _ := os.ReadDir(store.Dir)
	if len(entries) != 0 {
		t.Errorf("files created: %v", entries)
	}
}

func TestCreateReadOnlyDir(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions")
	}
	store := newDir(t)
	os.Chmod(store.Dir, 0o555)
	t.Cleanup(func() { os.Chmod(store.Dir, 0o755) })
	m, msg := press(New(store, nil, "foo"), "enter")
	if msg != nil || m.err == "" {
		t.Errorf("msg = %#v, err = %q", msg, m.err)
	}
}

func TestViewEmptyHint(t *testing.T) {
	v := New(newDir(t), nil, "").View()
	if !strings.Contains(v, "Type a name") && !strings.Contains(v, "No notes yet") {
		t.Errorf("missing hint:\n%s", v)
	}
	if !strings.Contains(v, "press Enter to create the first one") {
		t.Errorf("hint text wrong:\n%s", v)
	}
}

func TestViewWithCreateRow(t *testing.T) {
	v := New(newDir(t, "todos"), nil, "todo").View()
	lines := strings.Split(v, "\n")
	if !strings.Contains(lines[1], "> todos") {
		t.Errorf("line 1 = %q", lines[1])
	}
	if !strings.Contains(lines[2], `Create "todo"`) {
		t.Errorf("line 2 = %q", lines[2])
	}
	if strings.Contains(v, "No notes yet") {
		t.Error("hint shown for non-empty directory")
	}
}

// templatesDir makes a notes dir holding the given notes and a .templates
// subdirectory holding the given templates (later names are newer), and
// returns the notes store.
func templatesDir(t *testing.T, noteNames, tplNames []string) notes.Store {
	t.Helper()
	store := newDir(t, noteNames...)
	if tplNames == nil {
		return store
	}
	tdir := templates.Dir(store.Dir)
	if err := os.Mkdir(tdir, 0o755); err != nil {
		t.Fatal(err)
	}
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i, n := range tplNames {
		p := filepath.Join(tdir, n+".md")
		if err := os.WriteFile(p, []byte("# "+n+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		mt := base.Add(time.Duration(i) * time.Hour)
		os.Chtimes(p, mt, mt)
	}
	return store
}

func TestScopes(t *testing.T) {
	store := newDir(t, "a")
	if got := New(store, nil, "").Scope(); got != ScopeNotes {
		t.Errorf("New scope = %v", got)
	}
	if got := NewTemplates(templates.Store(store.Dir), "").Scope(); got != ScopeTemplates {
		t.Errorf("NewTemplates scope = %v", got)
	}
}

func TestTemplatesMissingDirIsEmpty(t *testing.T) {
	store := templatesDir(t, []string{"todo"}, nil)
	m := NewTemplates(templates.Store(store.Dir), "")
	if len(m.Rows()) != 0 || m.err != "" {
		t.Errorf("rows = %v err = %q", names(m.Rows()), m.err)
	}
	if v := m.View(); strings.Contains(v, "no such file") {
		t.Errorf("view shows an error:\n%s", v)
	}
}

func TestTemplatesListOnlyTemplates(t *testing.T) {
	store := templatesDir(t, []string{"todo"}, []string{"meeting"})
	m := NewTemplates(templates.Store(store.Dir), "")
	if got := names(m.Rows()); !reflect.DeepEqual(got, []string{"meeting"}) {
		t.Errorf("template rows = %v", got)
	}
	n := New(store, nil, "")
	if got := names(n.Rows()); !reflect.DeepEqual(got, []string{"todo"}) {
		t.Errorf("note rows = %v", got)
	}
}

func TestTemplatesNewestFirst(t *testing.T) {
	store := templatesDir(t, nil, []string{"old", "new"})
	m := NewTemplates(templates.Store(store.Dir), "")
	if got := names(m.Rows()); !reflect.DeepEqual(got, []string{"new", "old"}) {
		t.Errorf("rows = %v", got)
	}
}

func TestTemplateCreateMakesDirAndSkeleton(t *testing.T) {
	store := templatesDir(t, []string{"todo"}, nil)
	m := NewTemplates(templates.Store(store.Dir), "meeting")
	if got := names(m.Rows()); !reflect.DeepEqual(got, []string{"+meeting"}) {
		t.Fatalf("rows = %v", got)
	}
	m, msg := press(m, "enter")
	got, ok := msg.(NoteCreatedMsg)
	if !ok || got.Note.Name != "meeting" || got.Scope != ScopeTemplates {
		t.Fatalf("msg = %#v err = %q", msg, m.err)
	}
	b, err := os.ReadFile(filepath.Join(store.Dir, ".templates", "meeting.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != templates.Skeleton {
		t.Errorf("content = %q", b)
	}
	if _, err := templates.Parse(string(b)); err != nil {
		t.Errorf("skeleton does not parse: %v", err)
	}
	// The notes listing is untouched by the new directory.
	if got := names(New(store, nil, "").Rows()); !reflect.DeepEqual(got, []string{"todo"}) {
		t.Errorf("note rows = %v", got)
	}
}

func TestTemplateCreateExistingDir(t *testing.T) {
	store := templatesDir(t, nil, []string{"daily"})
	m := NewTemplates(templates.Store(store.Dir), "meeting")
	_, msg := press(m, "enter")
	if _, ok := msg.(NoteCreatedMsg); !ok {
		t.Fatalf("msg = %#v", msg)
	}
	if b, _ := os.ReadFile(filepath.Join(store.Dir, ".templates", "daily.md")); string(b) != "# daily\n" {
		t.Errorf("existing template touched: %q", b)
	}
}

func TestTemplateCreateInvalidName(t *testing.T) {
	store := templatesDir(t, nil, nil)
	m := NewTemplates(templates.Store(store.Dir), "a/b")
	m, msg := press(m, "enter")
	if msg != nil || m.err == "" || !strings.Contains(m.View(), m.err) {
		t.Errorf("msg = %#v, err = %q", msg, m.err)
	}
	if _, err := os.Stat(templates.Dir(store.Dir)); !os.IsNotExist(err) {
		t.Errorf(".templates created for an invalid name: %v", err)
	}
	entries, _ := os.ReadDir(store.Dir)
	if len(entries) != 0 {
		t.Errorf("files created: %v", entries)
	}
}

func TestCtrlTTogglesFromBothScopes(t *testing.T) {
	store := templatesDir(t, []string{"a"}, []string{"t"})
	if _, msg := press(New(store, nil, "q"), "ctrl+t"); msg != (ToggleScopeMsg{}) {
		t.Errorf("notes scope ctrl+t = %#v", msg)
	}
	if _, msg := press(NewTemplates(templates.Store(store.Dir), ""), "ctrl+t"); msg != (ToggleScopeMsg{}) {
		t.Errorf("template scope ctrl+t = %#v", msg)
	}
}

func TestTemplateScopeEsc(t *testing.T) {
	store := templatesDir(t, nil, []string{"t"})
	_, msg := press(NewTemplates(templates.Store(store.Dir), ""), "esc")
	if _, ok := msg.(BackMsg); !ok {
		t.Errorf("esc msg = %#v", msg)
	}
}

func TestTemplateScopeEnterOpens(t *testing.T) {
	store := templatesDir(t, nil, []string{"meeting"})
	_, msg := press(NewTemplates(templates.Store(store.Dir), ""), "enter")
	got, ok := msg.(OpenNoteMsg)
	if !ok || got.Note.Name != "meeting" || got.Scope != ScopeTemplates {
		t.Errorf("msg = %#v", msg)
	}
}

func TestTemplateViewPromptAndHints(t *testing.T) {
	store := templatesDir(t, nil, []string{"meeting"})
	v := NewTemplates(templates.Store(store.Dir), "").View()
	lines := strings.Split(v, "\n")
	if !strings.HasPrefix(stripANSI(lines[0]), "templates> ") {
		t.Errorf("prompt line = %q", lines[0])
	}
	if !strings.Contains(lines[1], "> meeting") {
		t.Errorf("line 1 = %q", lines[1])
	}
	last := lines[len(lines)-1]
	for _, want := range []string{"enter edit", "esc notes", "ctrl+t notes", "ctrl+o change directory"} {
		if !strings.Contains(last, want) {
			t.Errorf("footer %q lacks %q", last, want)
		}
	}
	if strings.Contains(v, "No templates yet") {
		t.Error("empty hint shown with a template present")
	}
	nv := New(store, nil, "").View()
	if !strings.Contains(nv, "ctrl+t templates") || strings.Contains(nv, "templates> ") {
		t.Errorf("notes view:\n%s", nv)
	}
}

func TestTemplateViewEmptyHint(t *testing.T) {
	store := templatesDir(t, []string{"todo"}, nil)
	v := NewTemplates(templates.Store(store.Dir), "").View()
	if !strings.Contains(v, "No templates yet") || !strings.Contains(v, "press Enter to create the first one") {
		t.Errorf("missing hint:\n%s", v)
	}
	if strings.Contains(v, "No notes yet") {
		t.Errorf("notes hint in template scope:\n%s", v)
	}
}

func TestCurrentNoteNotHiddenInTemplateScope(t *testing.T) {
	store := templatesDir(t, []string{"todo"}, []string{"todo"})
	cur := &notes.Note{Name: "todo"}
	if got := names(New(store, cur, "").Rows()); len(got) != 0 {
		t.Errorf("notes scope rows = %v, want current hidden", got)
	}
	m := NewTemplates(templates.Store(store.Dir), "")
	if got := names(m.Rows()); !reflect.DeepEqual(got, []string{"todo"}) {
		t.Errorf("template rows = %v", got)
	}
	if !strings.Contains(m.View(), "> todo") {
		t.Errorf("view lacks the template:\n%s", m.View())
	}
}

var ansi = regexp.MustCompile("\x1b\\[[0-9;]*m")

func stripANSI(s string) string { return ansi.ReplaceAllString(s, "") }

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
