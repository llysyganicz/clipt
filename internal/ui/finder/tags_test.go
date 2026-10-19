package finder

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/llysyganicz/clipt/internal/frontmatter"
	"github.com/llysyganicz/clipt/internal/templates"
	"github.com/llysyganicz/clipt/internal/ui/uitest"
)

// writeAll overwrites existing files in dir with the given contents.
func writeAll(t *testing.T, dir string, contents map[string]string) {
	t.Helper()
	for n, c := range contents {
		if err := os.WriteFile(filepath.Join(dir, n+".md"), []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// rowLine returns the stripped view line showing name.
func rowLine(t *testing.T, view, name string) string {
	t.Helper()
	for _, l := range strings.Split(uitest.Strip(view), "\n") {
		if strings.HasPrefix(strings.TrimLeft(l, "> "), name) {
			return l
		}
	}
	t.Fatalf("no row for %q in:\n%s", name, uitest.Strip(view))
	return ""
}

func sized(m Model, w int) Model {
	m, _ = m.Update(tea.WindowSizeMsg{Width: w, Height: 24})
	return m
}

func TestRowsShowTags(t *testing.T) {
	store := newDir(t, "ideas", "standup")
	writeAll(t, store.Dir, map[string]string{"standup": "---\ntags: [work, meeting]\n---\n"})
	v := sized(New(store, nil, ""), 80).View()
	if got := rowLine(t, v, "standup"); got != "> standup #work #meeting" {
		t.Errorf("tagged row = %q", got)
	}
	if got := rowLine(t, v, "ideas"); got != "  ideas" {
		t.Errorf("untagged row = %q", got)
	}
}

func TestRowsNarrowTerminal(t *testing.T) {
	store := newDir(t, "standup")
	writeAll(t, store.Dir, map[string]string{"standup": "---\ntags: [work, meeting]\n---\n"})
	// "> standup" is 9 cells, plus a space leaves 5 for the tags.
	if got := rowLine(t, sized(New(store, nil, ""), 15).View(), "standup"); got != "> standup #wor…" {
		t.Errorf("narrow row = %q", got)
	}
	if got := rowLine(t, sized(New(store, nil, ""), 11).View(), "standup"); got != "> standup" {
		t.Errorf("no-room row = %q", got)
	}
	if got := rowLine(t, sized(New(store, nil, ""), 5).View(), "standup"); got != "> standup" {
		t.Errorf("name must not be cut: %q", got)
	}
}

func TestTemplateRowsShowTags(t *testing.T) {
	store := templatesDir(t, nil, []string{"daily", "meeting"})
	writeAll(t, templates.Dir(store.Dir), map[string]string{"meeting": templates.SkeletonWithTags([]string{"meeting"})})
	m := sized(NewTemplates(templates.Store(store.Dir), ""), 80)
	if got := rowLine(t, m.View(), "meeting"); got != "> meeting #meeting" {
		t.Errorf("tagged row = %q", got)
	}
	if got := rowLine(t, m.View(), "daily"); got != "  daily" {
		t.Errorf("untagged row = %q", got)
	}
	if got := rowLine(t, sized(NewTemplates(templates.Store(store.Dir), ""), 14).View(), "meeting"); got != "> meeting #me…" {
		t.Errorf("narrow row = %q", got)
	}
}

func TestCreateRowShowsTags(t *testing.T) {
	v := sized(New(newDir(t), nil, "#work standup"), 80).View()
	if got := rowLine(t, v, `Create "standup"`); got != `> Create "standup" #work` {
		t.Errorf("create row = %q", got)
	}
}

func TestFilterByTag(t *testing.T) {
	store := newDir(t, "groceries", "standup", "draft")
	writeAll(t, store.Dir, map[string]string{
		"standup":   "---\ntags: [work]\n---\n",
		"groceries": "---\ntags: [home]\n---\n",
		"draft":     "---\ntags: [work]\n# never closed\n",
	})
	if got := names(New(store, nil, "#work").Rows()); !reflect.DeepEqual(got, []string{"standup"}) {
		t.Errorf("#work rows = %v", got)
	}
	m := New(store, nil, "")
	if got := names(m.Rows()); len(got) != 3 { // rewriting reset the mtimes
		t.Errorf("all rows = %v", got)
	}
	if got := strings.TrimLeft(rowLine(t, m.View(), "draft"), "> "); m.err != "" || got != "draft" {
		t.Errorf("broken front matter: err=%q row=%q", m.err, rowLine(t, m.View(), "draft"))
	}
}

func TestFilterTemplatesByTag(t *testing.T) {
	store := templatesDir(t, nil, []string{"daily", "meeting"})
	writeAll(t, templates.Dir(store.Dir), map[string]string{"meeting": "---\ntags: [meeting]\n---\n"})
	if got := names(NewTemplates(templates.Store(store.Dir), "#meet").Rows()); !reflect.DeepEqual(got, []string{"meeting"}) {
		t.Errorf("rows = %v", got)
	}
}

func TestTagsRereadOnEntry(t *testing.T) {
	store := newDir(t, "ideas")
	if got := names(New(store, nil, "#work").Rows()); len(got) != 0 {
		t.Fatalf("rows before = %v", got)
	}
	writeAll(t, store.Dir, map[string]string{"ideas": "---\ntags: [work]\n---\n"})
	if got := names(New(store, nil, "#work").Rows()); !reflect.DeepEqual(got, []string{"ideas"}) {
		t.Errorf("rows after = %v", got)
	}
}

func TestCreateWithTags(t *testing.T) {
	store := newDir(t)
	_, msg := press(New(store, nil, "#work #meeting #Work standup"), "enter")
	got, ok := msg.(NoteCreatedMsg)
	if !ok || got.Note.Name != "standup" || !reflect.DeepEqual(got.Note.Tags, []string{"work", "meeting"}) {
		t.Fatalf("msg = %#v", msg)
	}
	b, err := os.ReadFile(filepath.Join(store.Dir, "standup.md"))
	if err != nil || string(b) != "---\ntags: [work, meeting]\n---\n" {
		t.Errorf("content = %q, %v", b, err)
	}
}

func TestCreateWithoutTagsIsEmpty(t *testing.T) {
	store := newDir(t)
	if _, msg := press(New(store, nil, "plain"), "enter"); msg == nil {
		t.Fatal("no message")
	}
	if b, _ := os.ReadFile(filepath.Join(store.Dir, "plain.md")); len(b) != 0 {
		t.Errorf("content = %q", b)
	}
}

func TestCreateTagsWrittenAsTyped(t *testing.T) {
	store := newDir(t, "review")
	writeAll(t, store.Dir, map[string]string{"review": "---\ntags: [work]\n---\n"})
	m := New(store, nil, "#wo standup")
	m, _ = press(m, "down") // past review to the create row
	if _, msg := press(m, "enter"); msg == nil {
		t.Fatal("no message")
	}
	b, _ := os.ReadFile(filepath.Join(store.Dir, "standup.md"))
	tags, err := frontmatter.Parse(string(b))
	if err != nil || !reflect.DeepEqual(tags, []string{"wo"}) {
		t.Errorf("tags = %q, %v", tags, err)
	}
}

func TestCreateTemplateWithTags(t *testing.T) {
	store := newDir(t)
	_, msg := press(NewTemplates(templates.Store(store.Dir), "#meeting standup"), "enter")
	if got, ok := msg.(NoteCreatedMsg); !ok || got.Scope != ScopeTemplates {
		t.Fatalf("msg = %#v", msg)
	}
	b, err := os.ReadFile(filepath.Join(templates.Dir(store.Dir), "standup.md"))
	if err != nil || string(b) != templates.SkeletonWithTags([]string{"meeting"}) {
		t.Errorf("content = %q, %v", b, err)
	}
	tpl, err := templates.Parse(string(b))
	if err != nil || !reflect.DeepEqual(tpl.Tags, []string{"meeting"}) {
		t.Errorf("parsed tags = %q, %v", tpl.Tags, err)
	}
}

func TestCreateTemplateWithoutTags(t *testing.T) {
	store := newDir(t)
	press(NewTemplates(templates.Store(store.Dir), "daily"), "enter")
	b, _ := os.ReadFile(filepath.Join(templates.Dir(store.Dir), "daily.md"))
	if string(b) != templates.Skeleton || !strings.Contains(string(b), "tags: []") {
		t.Errorf("content = %q", b)
	}
}
